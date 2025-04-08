package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whit-colm/itsc-4155-project/internal/testhelper/dummyvalues"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

// Useful to check that a type implements an interface
var _ repository.DummyPopulator = (*bookRepository)(nil)

func (b *bookRepository) PopulateDummyValues(ctx context.Context) error {
	batch := &pgx.Batch{}
	// TODO: find a way to *not* make this like. O(N*M)??
	for _, book := range dummyvalues.ExampleBooks {
		batch.Queue(`INSERT INTO books (id, title, published)
					 VALUES ($1, $2, $3)`,
			book.ID, book.Title, book.Published.In(time.UTC))
		// isbn used instead of i because `i` generally means index
		for _, isbn := range book.ISBNs {
			batch.Queue(`INSERT INTO isbns (isbn, book_id, isbn_type)
					 VALUES ($1, $2, $3)
					 ON CONFLICT (isbn) DO NOTHING`,
				isbn.String(), book.ID, isbn.Version().String())
		}
	}

	results := b.db.SendBatch(ctx, batch)
	defer results.Close()

	// we can use results.Exec() to iterate through each query, sorta
	// like a stack of results from my understanding (or queue, w/e)
	// the transaction is already done though; this is just to check
	// for errors (TODO: loop?)
	/*
		ct, err := results.Exec()
		if err != nil {
			fmt.Errorf("error executing batch: %w", err)
		}
		if c
	*/

	return nil
}

func (b *bookRepository) IsPrepopulated(ctx context.Context) bool {
	var ids uuid.UUIDs
	for _, v := range dummyvalues.ExampleBooks {
		ids = append(ids, v.ID)
	}
	var count int
	if err := b.db.QueryRow(ctx,
		`SELECT 
			 COUNT(*)
		 FROM books b
		 WHERE b.id = any($1)`,
		ids).Scan(&count); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return false
	}
	return count == len(ids)
}

func (b *bookRepository) CleanDummyValues(ctx context.Context) error {
	tx, err := b.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err = b.db.Exec(ctx, `TRUNCATE books CASCADE`); err != nil {
		return fmt.Errorf("could not truncate: %w", err)
	}
	return tx.Commit(ctx)
}

func (a *bookRepository) SetDatastore(ctx context.Context, ds any) error {
	if db, ok := ds.(*pgxpool.Pool); !ok {
		return fmt.Errorf("unable to cast ds into pgxpool Pool.")
	} else {
		a.db = db
		return nil
	}
}

/** Actual tests below **/

func TestBookCreate(t *testing.T) {
	if err := br.Create(t.Context(), &dummyvalues.ExampleBook); err != nil {
		t.Errorf("could not create book `%v`: %s", dummyvalues.ExampleBook, err)
	}
}

func TestBookDelete(t *testing.T) {
	if err := br.Delete(t.Context(), dummyvalues.ExampleBook.ID); err != nil {
		t.Errorf("could not delete book `%v`: %s", dummyvalues.ExampleBook, err)
	}
}

func TestBookGetByID(t *testing.T) {
	if b, err := br.GetByID(t.Context(), dummyvalues.ExampleBooks[0].ID); err != nil {
		t.Errorf("error finding known UUID: %s", err)
	} else if !dummyvalues.IsBookEquals(*b, dummyvalues.ExampleBooks[0]) {
		t.Errorf("inequality between fetched and known book: want %v; have %v", *b, dummyvalues.ExampleBook)
	}

	deadUUID, err := uuid.Parse("00000000-0000-8000-0000-200000000000")
	if err != nil {
		t.Errorf("Error generating dead UUID: %s", err)
	}
	if b, err := br.GetByID(t.Context(), deadUUID); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("unexpected error found for dead UUID: %s", err)
	} else if b != nil {
		t.Errorf("unexpected found book for dead UUID: %v", b)
	}
}

func TestBookGetByISBN(t *testing.T) {
	if b, err := br.GetByISBN(t.Context(), dummyvalues.ExampleBook.ISBNs[0]); err != nil {
		t.Errorf("error finding known ISBN: %s", err)
	} else if !dummyvalues.IsBookEquals(*b, dummyvalues.ExampleBook) {
		t.Errorf("inequality between fetched and known book: want %v; have %v", *b, dummyvalues.ExampleBook)
	}

	deadISBN := model.MustNewISBN("978-1408855652")
	if b, err := br.GetByISBN(t.Context(), deadISBN); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("unexpected error searching dead ISBN: %s", err)
	} else if b.ID != uuid.Nil {
		t.Errorf("unexpected found book for dead ISBN: %v", b)
	}
}

func TestBookSearch(t *testing.T) {
	bs, err := br.Search(t.Context())
	if err != nil {
		t.Errorf("unexpected error in search: %s", err)
	}
	if !dummyvalues.IsBookSliceEquals(bs, dummyvalues.ExampleBooks) {
		t.Errorf("unexpected inequality with fetched books")
	}
}
