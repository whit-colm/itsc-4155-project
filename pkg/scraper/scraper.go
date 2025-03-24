package scraper

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"cloud.google.com/go/civil"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
)

// URL for theGoogle Books API
const googleBooksAPI = "https://www.googleapis.com/books/v1/volumes?q="

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

// Detailed book information
type VolumeInfo struct {
	Title               string               `json:"title"`
	Authors             []string             `json:"authors"`
	PublishedDate       string               `json:"publishedDate"`
	IndustryIdentifiers []Identifier `json:"industryIdentifiers"`
}

// Representation of a book
type Book struct {
	ID        uuid.UUID  `json:"id"`        
	ISBNs     []model.ISBN     `json:"isbns"`     
	Title     string     `json:"title"`     
	Author    string     `json:"author"`   
	Published civil.Date `json:"published"` 
}

// Gets the book data using ISBN and converts it
func FetchBookByISBN(isbn string) (*model.Book, error) {
	url := fmt.Sprintf("%sisbn:%s", googleBooksAPI, isbn)

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

	book := &model.Book{
		Title: bookData.Title,
		Author: func() string {
			if len(bookData.Authors) > 0 {
				return bookData.Authors[0]
			}
			return "Unknown Author"
		}(),
		Published: published, 
		ISBNs: extractISBN(bookData.IndustryIdentifiers),
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


var db *sqlx.DB

// Connect to the database
func InitDB(dataSourceName string) error {
	var err error
	db, err = sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	return nil
}

// StoreBook saves the data into the database
func StoreBook(book *model.Book) error {
	var publishedTime time.Time
	if !book.Published.IsZero() {
		publishedTime = time.Date(
			book.Published.Year, 
			time.Month(book.Published.Month), 
			book.Published.Day, 
			0, 0, 0, 0, 
			time.UTC,
		)
	}

	query := `
	INSERT INTO books (id, title, author, published, isbn)
	VALUES ($1, $2, $3, $4, $5)
	ON CONFLICT (isbn) DO NOTHING;
	`

	isbn := ""
	if len(book.ISBNs) > 0 {
		isbn = book.ISBNs[0].String()
	}

	_, err := db.Exec(query, book.ID, book.Title, book.Author, publishedTime, isbn)
	if err != nil {
		return fmt.Errorf("failed to insert book: %v", err)
	}
	return nil
}
