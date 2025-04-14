package scraper

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/whit-colm/itsc-4155-project/internal/testhelper/mockdatastore"
)

// Test for FetchBookByISBN
func TestFetchBookByISBN(t *testing.T) {
	ctx := t.Context()
	mockBlobManager := mockdatastore.NewInMemoryRepository[string]().Blob
	isbn := "9780143127741"

	book, err := FetchBookByISBN(ctx, isbn, mockBlobManager)
	if err != nil {
		t.Fatalf("Error, got %v", err)
	}

	if book.Title == "" {
		t.Error("No Book title Found")
	} else {
		t.Logf("Parsed Title: %s", book.Title)
	}

	if book.AuthorIDs == nil {
		t.Error("No Authors Found")
	} else {
		t.Logf("Parsed Author IDs: %s", book.AuthorIDs)
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

// Test for storeLargeContent
func TestStoreLargeContent(t *testing.T) {
	ctx := context.Background()
	mockBlobManager := mockdatastore.NewInMemoryRepository[string]().Blob
	content := "This is a test description that would be over 2KB in real usage"

	id, err := storeLargeContent(ctx, content, mockBlobManager)
	assert.NoError(t, err, "Should not return error")
	assert.NotEqual(t, uuid.Nil, id, "Should return valid UUID")

	id, err = storeLargeContent(ctx, "", mockBlobManager)
	assert.NoError(t, err, "Empty content should not error")
	assert.Equal(t, uuid.UUID{}, id, "Empty content should return zero UUID")
}

// Test for getFirstAuthor
func TestGetFirstAuthor(t *testing.T) {
	assert.Equal(t, "John Doe", getFirstAuthor([]string{"John Doe", "Jane Smith"}), "Should return first author")
	assert.Equal(t, "Unknown Author", getFirstAuthor([]string{}), "Should return default")
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
		{"invalid-date", "0001-01-01"}, // zero date
	}

	for _, tt := range tests {
		date := parsePublishedDate(tt.input)
		assert.Equal(t, tt.expected, date.String(), "Failed parsing date: "+tt.input)
	}
}