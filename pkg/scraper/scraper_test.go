package scraper

import (
	"context"
	"testing"
	
	"github.com/stretchr/testify/assert"
	"github.com/google/uuid"
)

// MockBlobManagaer is a mock implementation of BlobManager for testing
type MockBlobManager struct{}

func (m *MockBlobManager) Store(ctx context.Context, id uuid.UUID, content []byte) error {
	return nil
}

func (m *MockBlobManager) Retrieve(ctx context.Context, id uuid.UUID) ([]byte, error) {
	return nil, nil
}

func (m *MockBlobManager) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

// Test for FetchBookByISBN
func TestFetchBookByISBN(t *testing.T) {
	ctx := context.Background()
	mockBlobManager := &MockBlobManager{}
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

// Test for storeLargeContent
func TestStoreLargeContent(t *testing.T) {
	ctx := context.Background()
	mockBlobManager := &MockBlobManager{}
	content := "This is a test description that would be over 2KB in real usage"

	id, err := storeLargeContent(ctx, content, mockBlobManager)
	assert.NoError(t, err, "Should not return error")
	assert.NotEqual(t, uuid.Nil, id, "Should return valid UUID")

	id, err = storeLargeContent(ctx, "", mockBlobManager)
	assert.NoError(t, err, "Empty content should not error")
	assert.Equal(t, uuid.UUID{}, id, "Empty content should return zero UUID")
}
