package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"context"
	"cloud.google.com/go/civil"
	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
)

// URL for theGoogle Books API
const (
	googleBooksAPI = "https://www.googleapis.com/books/v1/volumes?q=isbn:%s"
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
    Thumbnail string `json:"thumbnail"`
    Small     string `json:"small"`
    Medium    string `json:"medium"`
    Large     string `json:"large"`
    ExtraLarge string `json:"extraLarge"`
}

// Struct for main book information
type VolumeInfo struct {
	Title               string       `json:"title"`
	Authors             []string     `json:"authors"`
	PublishedDate       string       `json:"publishedDate"`
	Description         string       `json:"description"`
	IndustryIdentifiers []Identifier `json:"industryIdentifiers"`
	ImageLinks         	ImageLinks   `json:"imageLinks"`
}

// Gets the book data using ISBN and converts it
func FetchBookByISBN(ctx context.Context, isbn string, blobManager repository.BlobManager) (*model.Book, error) {
	url := fmt.Sprintf(googleBooksAPI, isbn)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	var googleResp GoogleBooksResponse
	if err := json.Unmarshal(body, &googleResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	if len(googleResp.Items) == 0 {
		return nil, fmt.Errorf("no book found for ISBN: %s", isbn)
	}

	bookData := googleResp.Items[0].VolumeInfo

	fmt.Printf("Parsed Title: %s\n", bookData.Title)
	fmt.Printf("Parsed Authors: %v\n", bookData.Authors)


	var published civil.Date
	if bookData.PublishedDate != "" {
		parsedTime, err := time.Parse("2006-01-02", bookData.PublishedDate)
		if err == nil {
			published = civil.DateOf(parsedTime)
		} else {
			fmt.Printf("Could not parse PublishedDate: %v\n", err)
		}
	}

	// Stores large description
	descriptionRef := uuid.UUID{}
	if len(bookData.Description) > maxContentSize {
		descriptionRef, err = storeLargeContent(ctx, bookData.Description, blobManager)
		if err != nil {
			return nil, err
		}
	}

	// Stores cover image
	imageRef := uuid.UUID{}
	if bookData.ImageLinks.Thumbnail != "" {
    	imageRef, err = storeImage(ctx, bookData.ImageLinks.Thumbnail, blobManager)
    	if err != nil {
       		fmt.Printf("Warning: failed to store image (continuing without): %v\n", err)    
    	}
	}

	// Create book model
	book := &model.Book{
		ID:        uuid.New(),
		Title:     bookData.Title,
		Author:    getFirstAuthor(bookData.Authors),
		Published: published,
		ISBNs:       extractISBN(bookData.IndustryIdentifiers),
		Description: descriptionRef,
		CoverImage:  imageRef,
	}

	return book, nil
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

// storeLargeContent saves large text content
func storeLargeContent(ctx context.Context, content string, blobManager repository.BlobManager) (uuid.UUID, error) {
	ref := uuid.New()
	if err := blobManager.Store(ctx, ref, []byte(content)); err != nil {
		return uuid.UUID{}, fmt.Errorf("failed to store large content: %v", err)
	}
	return ref, nil
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

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to read image data: %v", err)
	}

	ref := uuid.New()
	if err := blobManager.Store(ctx, ref, imageData); err != nil {
		return uuid.Nil, fmt.Errorf("failed to store image: %v", err)
	}

	return ref, nil
}

// StoreBook saves the data into the database
func StoreBook(ctx context.Context, book *model.Book, bookManager repository.BookManager) error {
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