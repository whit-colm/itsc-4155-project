package scraper

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

// URL for theGoogle Books API
const (
	maxContentSize = 2 * 1024
)

// Struct for the response from the API
type GoogleBooksResponse struct {
	Items []struct {
		VolumeInfo VolumeInfo `json:"volumeInfo"`
	} `json:"items"`
}

// Struct to hold identifiers
type Identifier struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}

// Struct for different image sizes
type ImageLinks struct {
	SmallThumbnail string `json:"smallThumbnail"`
	Thumbnail      string `json:"thumbnail"`
	Small          string `json:"small"`
	Medium         string `json:"medium"`
	Large          string `json:"large"`
	ExtraLarge     string `json:"extraLarge"`
}

// Struct for main book information
type VolumeInfo struct {
	Title               string       `json:"title"`
	Subtitle            string       `json:"subtitle"`
	Authors             []string     `json:"authors"`
	PublishedDate       string       `json:"publishedDate"`
	Description         string       `json:"description"`
	IndustryIdentifiers []Identifier `json:"industryIdentifiers"`
	ImageLinks          ImageLinks   `json:"imageLinks"`
}

// extractISBN extracts the ISBN
func extractISBN(identifiers []Identifier) []model.ISBN {
	var isbns []model.ISBN

	for _, id := range identifiers {
		if id.Type == "ISBN_13" {
			isbns = append(isbns, model.MustNewISBN(id.Identifier, model.ISBN13))
		}
	}

	for _, id := range identifiers {
		if id.Type == "ISBN_10" {
			isbns = append(isbns, model.MustNewISBN(id.Identifier, model.ISBN10))
		}
	}

	return isbns
}

func urlToBlob(ctx context.Context, imageURL string) (*model.Blob, error) {
	const errorCaller = "fetch url to blob"
	// TODO: Use gzip to compress the image in transit
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", errorCaller, err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", errorCaller, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%v: unexpected status code %d", errorCaller, resp.StatusCode)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return nil, fmt.Errorf("%v: %w", errorCaller, err)
	}

	blob := &model.Blob{
		ID:      id,
		Content: resp.Body,
		Metadata: map[string]string{
			"content-type": resp.Header.Get("Content-Type"),
			"source-url":   imageURL,
			"size":         resp.Header.Get("Content-Length"),
		},
	}

	return blob, nil
}

// storeImage downloads and stores an image
func storeImage(ctx context.Context, imageURL string, blobManager repository.BlobManager) (uuid.UUID, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", imageURL, nil)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create image request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to fetch image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return uuid.Nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	ref := uuid.New()
	blob := model.Blob{
		ID:      ref,
		Content: resp.Body,
		Metadata: map[string]string{
			"content-type": resp.Header.Get("Content-Type"),
			"source-url":   imageURL,
			"size":         resp.Header.Get("Content-Length"),
		},
	}

	if err := blobManager.Create(ctx, &blob); err != nil {
		return uuid.Nil, fmt.Errorf("failed to store image: %v", err)
	}

	return ref, nil
}

// StoreBook saves the data into the database
func StoreBook(ctx context.Context, book *model.Book, bookManager repository.BookManager[*model.Book]) error {
	if book == nil {
		return fmt.Errorf("book cannot be nil")
	}

	if book.Title == "" {
		return fmt.Errorf("book title cannot be empty")
	}

	err := bookManager.Create(ctx, book)
	if err != nil {
		return fmt.Errorf("failed to store book: %v", err)
	}
	return nil
}

// getFirstAuthor returns first author or "Unknown Author"
func getFirstAuthor(authors []string) string {
	if len(authors) > 0 {
		return authors[0]
	}
	return "Unknown Author"
}

// Converts a string date to civil.Date using different layouts
func parsePublishedDate(dateStr string) civil.Date {
	layouts := []string{"2006-01-02", "2006-01", "2006"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return civil.DateOf(t)
		}
	}
	fmt.Printf("Warning: could not parse date: %s\n", dateStr)
	return civil.Date{}
}

// Splits a full name into given and family name
func parseSingleAuthor(fullName string) *model.Author {
	fullName = strings.TrimSpace(fullName)
	if fullName == "" {
		return &model.Author{
			ID:         uuid.New(),
			GivenName:  "Unknown",
			FamilyName: "Author",
		}
	}

	if lastSpace := strings.LastIndex(fullName, " "); lastSpace != -1 {
		return &model.Author{
			ID:         uuid.New(),
			GivenName:  strings.TrimSpace(fullName[:lastSpace]),
			FamilyName: strings.TrimSpace(fullName[lastSpace+1:]),
		}
	}

	return &model.Author{
		ID:         uuid.New(),
		GivenName:  fullName,
		FamilyName: "",
	}

}

// Checks if ISBN is either 10 or 13 digits
func isValidISBN(isbn string) bool {
	cleanISBN := strings.ReplaceAll(strings.ReplaceAll(isbn, "-", ""), " ", "")

	if len(cleanISBN) != 10 && len(cleanISBN) != 13 {
		return false
	}

	return true
}
