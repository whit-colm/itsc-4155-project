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
	"github.com/whit-colm/itsc-4155-project/internal/testhelper"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
)

var br bookRepository

/// DummyPopulator implementations ///

func (b *bookRepository) PopulateDummyValues(ctx context.Context) error {
	batch := &pgx.Batch{}
	// TODO: find a way to *not* make this like. O(N*M)??
	for _, book := range testhelper.ExampleBooks {
		batch.Queue(`INSERT INTO books (id, title, author, published)
					 VALUES ($1, $2, $3, $4)`,
			book.ID, book.Title, book.AuthorID, book.Published.In(time.UTC))
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
	for _, v := range testhelper.ExampleBooks {
		ids = append(ids, v.ID)
	}
	var count int
	if err := b.db.QueryRow(ctx,
		`SELECT 
			 COUNT(*)
		 FROM books b
		 WHERE b.id = any($1)`,
		ids).Scan(&count); err != nil {
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

	var ids uuid.UUIDs = []uuid.UUID{testhelper.ExampleBook.ID}
	for _, v := range testhelper.ExampleBooks {
		ids = append(ids, v.ID)
	}

	if _, err := tx.Exec(ctx,
		`DELETE FROM books b
		 WHERE b.id = any($1)`,
		ids); err != nil {
		return fmt.Errorf("failed to delete books: %w", err)
	}

	return tx.Commit(ctx)
}

// I was going to do this with flags, but tests and flags are real
// gnarly together, so we instead pass an env var. Yes I know this has
// Issues:tm: with Windows
func TestMain(m *testing.M) {
	// Not all devices are equipped to test the DB code, and that's ok
	// (well not really, but we're too poor and strapped for time to do
	// anything). So we check a db uri from ENV vars, if one exists we
	// use it and fail tests; otherwise we "pass" and skip the whole
	// sordid affair.
	uriString := os.Getenv("DB_URI")
	if uriString == "" {
		fmt.Printf("skipping tests; empty `DB_URI` variable.\n")
		os.Exit(0)
	}
	c := &postgres{}

	err := c.Connect(uriString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to PostgreSQL: %v\n", err)
		os.Exit(1)
	}
	defer c.Disconnect()
	br.db = c.db

	p := false
	if ctx := context.Background(); !br.IsPrepopulated(ctx) {
		if err := br.PopulateDummyValues(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to instantiate dummy values: %v\n", err)
			os.Exit(1)
		}
		p = true
		defer br.CleanDummyValues(ctx)
	} else {
		br.CleanDummyValues(context.Background())
		fmt.Fprintf(os.Stderr, "Database is already populated\n")
		os.Exit(1)
	}
	code := m.Run()

	// This was going to be a defer but it doesn't work for some reason.
	if p {
		if err := br.CleanDummyValues(context.Background()); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to clean dummy values: %v\n", err)
			os.Exit(1)
		}
	}

	os.Exit(code)
}

func TestPing(t *testing.T) {
	if err := br.db.Ping(context.Background()); err != nil {
		t.Errorf("database failed ping: %s", err)
	}
}

func TestCreate(t *testing.T) {
	if err := br.Create(t.Context(), &testhelper.ExampleBook); err != nil {
		t.Errorf("could not create book `%v`: %s", testhelper.ExampleBook, err)
	}
}

func TestGetByID(t *testing.T) {
	if b, err := br.GetByID(t.Context(), testhelper.ExampleBook.ID); err != nil {
		t.Errorf("error finding known UUID: %s", err)
	} else if !testhelper.IsBookEquals(*b, testhelper.ExampleBook) {
		t.Errorf("inequality between fetched and known book: want %v; have %v", *b, testhelper.ExampleBook)
	}

	deadUUID, err := uuid.NewV7()
	if err != nil {
		t.Errorf("Error generating dead UUID: %s", err)
	}
	if b, err := br.GetByID(t.Context(), deadUUID); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("unexpected error found for dead UUID: %s", err)
	} else if b != nil {
		t.Errorf("unexpected found book for dead UUID: %v", b)
	}
}

func TestGetByISBN(t *testing.T) {
	if _, b, err := br.GetByISBN(t.Context(), testhelper.ExampleBook.ISBNs[0]); err != nil {
		t.Errorf("error finding known ISBN: %s", err)
	} else if !testhelper.IsBookEquals(*b, testhelper.ExampleBook) {
		t.Errorf("inequality between fetched and known book: want %v; have %v", *b, testhelper.ExampleBook)
	}

	deadISBN := model.MustNewISBN("978-1408855652")
	if i, _, err := br.GetByISBN(t.Context(), deadISBN); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("unexpected error searching dead ISBN: %s", err)
	} else if i != uuid.Nil {
		t.Errorf("unexpected found book for dead ISBN: %v", i)
	}
}

func TestSearch(t *testing.T) {
	bs, err := br.Search(t.Context())
	if err != nil {
		t.Errorf("unexpected error in search: %s", err)
	}
	if !testhelper.IsBookSliceEquals(bs, append(testhelper.ExampleBooks, testhelper.ExampleBook)) {
		t.Errorf("unexpected inequality with fetched books")
	}
}

func TestDelete(t *testing.T) {
	if err := br.Delete(t.Context(), &testhelper.ExampleBook); err != nil {
		t.Errorf("could not delete book `%v`: %s", testhelper.ExampleBook, err)
	}
}
