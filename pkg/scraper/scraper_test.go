package scraper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/whit-colm/itsc-4155-project/internal/testhelper/dummyvalues"
	"github.com/whit-colm/itsc-4155-project/internal/testhelper/mockdatastore"
)

// Test for FetchBookByISBN
func TestFetchBookByISBN(t *testing.T) {
	ctx := t.Context()
	repo := mockdatastore.NewInMemoryRepository[string]()
	scrp := NewBookScraper(repo.Blob, repo.Book, repo.Author)

	isbn := dummyvalues.ExampleBook.ISBNs[0]

	count, err := scrp.ScrapeISBN(ctx, isbn)
	if err != nil {
		t.Fatalf("Error, got %v", err)
	}

	assert.Equal(t, count, 1, "One book should be found")
	book, err := repo.Book.GetByISBN(ctx, isbn)
	if err != nil {
		t.Fatalf("Error fetching book by ISBN: %v", err)
	}
	assert.NotNil(t, book, "Book should not be nil")
}

// Test for extractISBN
func TestExtractISBN(t *testing.T) {
	identifiers := []industryIdentifier{
		{Type: "ISBN_13", Identifier: "9783161484100"},
		{Type: "ISBN_10", Identifier: "316148410X"},
	}

	isbns := extractISBN(identifiers)

	assert.Len(t, isbns, 2, "Expected 2 ISBNs, got %d", len(isbns))
	assert.Equal(t, "9783161484100", isbns[0].String(), "First ISBN mismatch")
	assert.Equal(t, "316148410X", isbns[1].String(), "Second ISBN mismatch")
}

// Test for parsePublishedDate
func TestParsePublishedDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2006-01-02", "2006-01-02"},
		{"2006-01", "2006-01-01"},
		{"2006", "2006-01-01"},
		{"invalid-date", "0000-00-00"},
	}

	for _, tt := range tests {
		date := parsePublishedDate(tt.input)
		assert.Equal(t, tt.expected, date.String(), "Failed parsing date: "+tt.input)
	}
}
