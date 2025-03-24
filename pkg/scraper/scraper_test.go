package scraper

import (
	"testing"

	"cloud.google.com/go/civil"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
)


// Test for FetchBookByISBN
func TestFetchBookByISBN(t *testing.T) {
	isbn := "9780143127741" 

	book, err := FetchBookByISBN(isbn)
	if err != nil {
		t.Fatalf("Error, got %v", err)
	}

	if book.Title == "" {
		t.Error("No Book title Found")
	} else {
		t.Logf("Parsed Title: %s", book.Title)
	}

	if book.Author == "" {
		t.Error("No Author Found")
	} else {
		t.Logf("Parsed Author: %s", book.Author)
	}

	if len(book.ISBNs) == 0 {
		t.Error("No ISBN")
	} else {
		t.Logf("Parsed ISBNs: %v", book.ISBNs)
	}

	if book.Published.IsZero() {
		t.Error("No Published Date Found")
	} else {
		t.Logf("Parsed Published Date: %v", book.Published)
	}
}

// Test for extractISBN
func TestExtractISBN(t *testing.T) {
	identifiers := []Identifier{
		{Type: "ISBN_13", Identifier: "9783161484100"},
		{Type: "ISBN_10", Identifier: "316148410X"},
	}

	isbns := extractISBN(identifiers)

	assert.Len(t, isbns, 2, "Expected 2 ISBNs, got %d", len(isbns))
	assert.Equal(t, "9783161484100", isbns[0].String(), "First ISBN mismatch")
	assert.Equal(t, "316148410X", isbns[1].String(), "Second ISBN mismatch")
}

// Test to setup and initialize the database
func setupTestDB(t *testing.T) {
	var err error
	db, err = sqlx.Connect("sqlite3", ":memory:") 
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}

	schema := `
	CREATE TABLE books (
		id UUID PRIMARY KEY,
		title TEXT,
		author TEXT,
		published DATE,
		isbn TEXT UNIQUE
	);`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("Failed to create schema: %v", err)
	}
}

// Test to store a book and check if it is correctly inserted
func TestStoreBook(t *testing.T) {
	setupTestDB(t)

	book := &model.Book{
		ID:        uuid.New(),
		Title:     "Test Book",
		Author:    "John Doe",
		Published: civil.Date{Year: 2020, Month: 1, Day: 15},
		ISBNs:     []model.ISBN{model.MustNewISBN("9783161484100", model.ISBN13)},
	}

	if err := StoreBook(book); err != nil {
		t.Fatalf("Failed to store book: %v", err)
	}

	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM books WHERE title = ?", book.Title)
	if err != nil {
		t.Fatalf("Failed to query database: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 book, found %d", count)
	}
}
