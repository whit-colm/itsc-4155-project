package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
)

var br bookRepository

/// Some dummy data to test books with ///

var exampleBooks []model.Book = []model.Book{
	{
		ID: uuid.MustParse("0124e053-3580-7000-8794-db4a97089840"),
		ISBNs: []model.ISBN{
			model.MustNewISBN("0141439602", model.ISBN10),
			model.MustNewISBN("9780141439600", model.ISBN13),
		},
		Title:     "A Tale of Two Cities",
		Author:    "Charles Dickens",
		Published: civil.Date{Year: 1859, Month: time.November, Day: 26},
	}, {
		ID: uuid.MustParse("0124e053-3580-7000-875a-c17e9ba5023c"),
		ISBNs: []model.ISBN{
			model.MustNewISBN("0156012197", model.ISBN10),
			model.MustNewISBN("9780156012195", model.ISBN13),
		},
		Title:     "The Little Prince",
		Author:    "Antoine de Saint-Exupéry",
		Published: civil.Date{Year: 1943, Month: time.April},
	}, {
		ID: uuid.MustParse("0124e053-3580-7000-9127-dd33bb29c893"),
		ISBNs: []model.ISBN{
			model.MustNewISBN("0062315005", model.ISBN10),
			model.MustNewISBN("9780061122415", model.ISBN13),
		},
		Title:     "The Alchemist",
		Author:    "Paulo Coelho",
		Published: civil.Date{Year: 1988},
	},
}

var testBook model.Book = model.Book{
	ID: uuid.MustParse("0124e053-3580-7000-a59a-fb9e45afdc80"),
	ISBNs: []model.ISBN{
		model.MustNewISBN("0062073486", model.ISBN10),
		model.MustNewISBN("978-0062073488", model.ISBN13),
	},
	Title:     "And Then There Were None",
	Author:    "Agatha Christie",
	Published: civil.Date{Year: 1939, Month: time.November, Day: 6},
}

/// DummyPopulator implementations ///

func (b *bookRepository) PopulateDummyValues(ctx context.Context) error {
	batch := &pgx.Batch{}
	// TODO: find a way to *not* make this like. O(N*M)??
	for _, book := range exampleBooks {
		batch.Queue(`INSERT INTO books (id, title, author, published)
					 VALUES ($1, $2, $3, $4)`,
			book.ID, book.Title, book.Author, book.Published.In(time.UTC))
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
	for _, v := range exampleBooks {
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

	var ids uuid.UUIDs = []uuid.UUID{testBook.ID}
	for _, v := range exampleBooks {
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

	c, err := New(uriString)
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
	if err := br.Create(t.Context(), &testBook); err != nil {
		t.Errorf("could not create book `%v`: %s", testBook, err)
	}
}

func TestGetByID(t *testing.T) {
	if b, err := br.GetByID(t.Context(), testBook.ID); err != nil {
		t.Errorf("error finding known UUID: %s", err)
	} else if b.ID != testBook.ID || /* Yes this is evil, I miss Rust */
		b.Author != testBook.Author ||
		b.Title != testBook.Title ||
		b.Published != testBook.Published ||
		// TODO: This is an actually unwell way to do it. O(N^2).
		func(s1, s2 []model.ISBN) bool {
			if len(s1) != len(s2) {
				return false
			}
			for _, v1 := range s1 {
				for _, v2 := range s2 {
					if !reflect.DeepEqual(v1, v2) {
						return false
					}
				}
			}
			return true
		}(b.ISBNs, testBook.ISBNs) {
		t.Errorf("inequality between fetched and known book: want %v; have %v", *b, testBook)
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
	if _, b, err := br.GetByISBN(t.Context(), testBook.ISBNs[0]); err != nil {
		t.Errorf("error finding known ISBN: %s", err)
	} else if b.ID != testBook.ID || /* Yes this is evil, I miss Rust */
		b.Author != testBook.Author ||
		b.Title != testBook.Title ||
		b.Published != testBook.Published ||
		// TODO: This is an actually unwell way to do it. O(N^2).
		func(s1, s2 []model.ISBN) bool {
			if len(s1) != len(s2) {
				return false
			}
			for _, v1 := range s1 {
				for _, v2 := range s2 {
					if !reflect.DeepEqual(v1, v2) {
						return false
					}
				}
			}
			return true
		}(b.ISBNs, testBook.ISBNs) {
		t.Errorf("inequality between fetched and known book: want %v; have %v", *b, testBook)
	}

	deadISBN := model.MustNewISBN("978-1408855652")
	if i, _, err := br.GetByISBN(t.Context(), deadISBN); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("unexpected error searching dead ISBN: %s", err)
	} else if i != uuid.Nil {
		t.Errorf("unexpected found book for dead ISBN: %v", i)
	}
}

func TestSearch(t *testing.T) {
	if err := br.Delete(t.Context(), &testBook); err != nil {
		t.Errorf("could not delete book `%v`: %s", testBook, err)
	}
}
