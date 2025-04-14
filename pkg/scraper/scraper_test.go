package scraper

import (
	"context"
	"testing"
	"net/http/httptest"
	"net/http"

	"github.com/google/uuid"
	"cloud.google.com/go/civil"
	"github.com/stretchr/testify/assert"

	"github.com/whit-colm/itsc-4155-project/pkg/model"
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
    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, id)

    id, err = storeLargeContent(ctx, "", mockBlobManager)
    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, id)
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
		{"invalid-date", "0000-00-00"},
	}

	for _, tt := range tests {
		date := parsePublishedDate(tt.input)
		assert.Equal(t, tt.expected, date.String(), "Failed parsing date: "+tt.input)
	}
}

func TestIsValidISBN(t *testing.T) {
    tests := []struct {
        name     string
        isbn     string
        expected bool
    }{
        {"Valid ISBN-10", "0123456789", true},
        {"Valid ISBN-10 with X", "012345678X", true},
        {"Valid ISBN-13", "9780123456789", true},
        {"Invalid length", "12345", false},
        {"With hyphens", "978-0-12-345678-9", true},
        {"Empty string", "", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := isValidISBN(tt.isbn)
            assert.Equal(t, tt.expected, result)
        })
    }
}

func TestParseSingleAuthor(t *testing.T) {
    tests := []struct {
        name           string
        fullName       string
        expectedGiven  string
        expectedFamily string
    }{
        {"Standard name", "John Doe", "John", "Doe"},
        {"Multiple given names", "John Michael Doe", "John Michael", "Doe"},
        {"Single name", "Madonna", "Madonna", ""},
        {"Empty name", "", "Unknown", "Author"},
        {"Extra spaces", "  John   Doe  ", "John", "Doe"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            author := parseSingleAuthor(tt.fullName)
            assert.Equal(t, tt.expectedGiven, author.GivenName)
            assert.Equal(t, tt.expectedFamily, author.FamilyName)
            assert.NotEqual(t, uuid.Nil, author.ID)
        })
    }
}

func TestStoreImage(t *testing.T) {
    ctx := context.Background()
    mockBlobManager := mockdatastore.NewInMemoryRepository[string]().Blob

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "image/jpeg")
        w.Write([]byte("mock image data"))
    }))
    defer ts.Close()

    t.Run("Success case", func(t *testing.T) {
        id, err := storeImage(ctx, ts.URL, mockBlobManager)
        assert.NoError(t, err)
        assert.NotEqual(t, uuid.Nil, id)
    })

    t.Run("Invalid URL", func(t *testing.T) {
        _, err := storeImage(ctx, "http://invalid.url", mockBlobManager)
        assert.Error(t, err)
    })

    t.Run("Non-200 status", func(t *testing.T) {
        failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            w.WriteHeader(http.StatusNotFound)
        }))
        defer failServer.Close()

        _, err := storeImage(ctx, failServer.URL, mockBlobManager)
        assert.Error(t, err)
    })
}

func TestStoreBook(t *testing.T) {
    ctx := context.Background()
    mockBookManager := mockdatastore.NewInMemoryRepository[*model.Book]()
    
    validISBN, err := model.NewISBN("978-3-16-148410-0", model.ISBN13)
    if err != nil {
        t.Fatalf("Failed to create test ISBN: %v", err)
    }

    testBook := &model.Book{
        ID:          uuid.New(),
        Title:       "Test Book",
        AuthorIDs:   []uuid.UUID{uuid.New()},
        Published:   civil.Date{Year: 2020, Month: 1, Day: 1},
        ISBNs:       []model.ISBN{validISBN},
    }

    t.Run("Success case", func(t *testing.T) {
        err := StoreBook(ctx, testBook, mockBookManager.Book)
        assert.NoError(t, err)
        storedBook, err := mockBookManager.Book.GetByID(ctx, testBook.ID)
        assert.NoError(t, err)
        assert.Equal(t, testBook.Title, storedBook.Title)
    })

    t.Run("Nil book", func(t *testing.T) {
        err := StoreBook(ctx, nil, mockBookManager.Book)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "book cannot be nil")
    })

    t.Run("Empty title", func(t *testing.T) {
        badBook := *testBook
        badBook.Title = ""
        err := StoreBook(ctx, &badBook, mockBookManager.Book)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "book title cannot be empty")
    })
}