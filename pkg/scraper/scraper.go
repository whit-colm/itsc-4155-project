package scraper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

const (
	googleBooksAPI string = "https://www.googleapis.com/books/v1/volumes?&orderBy=relevance&maxResults=%d&startIndex=%d&q=%s"
	maxLimit       int    = 40
)

type BookScraper struct {
	blob repository.BlobManager
	book repository.BookManager[*model.Book]
	athr repository.AuthorManager[*model.Author]
}

func NewBookScraper(blob repository.BlobManager, book repository.BookManager[*model.Book], athr repository.AuthorManager[*model.Author]) *BookScraper {
	return &BookScraper{
		blob: blob,
		book: book,
		athr: athr,
	}
}

func (s *BookScraper) scrapeGoogleBooks(ctx context.Context, offset, limit int, query string, iCh chan<- int, eCh chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	const errorCaller string = "Google Books Scraper"
	if limit > maxLimit {
		eCh <- fmt.Errorf("%s: limit %d is greater than %d, which is the maximum allowed by Google Books API",
			errorCaller, limit, maxLimit,
		)
		return
	}
	url := fmt.Sprintf(googleBooksAPI, limit, offset, query)
	resp, err := http.Get(url)
	if err != nil {
		eCh <- fmt.Errorf("%v: %w", errorCaller, err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		// If we hit the rate limit, we wait for half a second to a second
		// before retrying. This is a simple backoff strategy.
		// This is not a perfect solution, but it should work for most cases.
		randDelay := 500 + rand.Intn(500)
		time.Sleep(time.Millisecond * time.Duration(randDelay))
		s.scrapeGoogleBooks(ctx, offset, limit, query, iCh, eCh, wg)
		return
	} else if resp.StatusCode != http.StatusOK {
		eCh <- fmt.Errorf("%s: received status code %d from Google Books API", errorCaller, resp.StatusCode)
		return
	}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		eCh <- fmt.Errorf("%v: %w", errorCaller, err)
		return
	}

	var aux struct {
		Total int `json:"totalItems"`
		Items []struct {
			SelfLink string `json:"selfLink"`
		} `json:"items"`
	}
	if err := json.Unmarshal(respBody, &aux); err != nil {
		eCh <- fmt.Errorf("%v: %w", errorCaller, err)
		return
	} else if aux.Total == 0 {
		iCh <- -1 // No books found
		return
	}

	/*** Prepare request for each self-link ***/
	//const fields string = "fields=volumeInfo(title,subtitle,authors,publishedDate,description,industryIdentifiers,categories,imageLinks)"
	var links []string = make([]string, len(aux.Items))
	for i, v := range aux.Items {
		links[i] = v.SelfLink /*+ "&" + fields*/
	}

	newBook := func(link string) (int, error) {
		const errorCaller string = "self-link scrape"
		resp, err := http.Get(link)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", errorCaller, err)
		}
		defer resp.Body.Close()

		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", errorCaller, err)
		}
		var aux struct {
			VolumeInfo struct {
				Title               string       `json:"title"`
				Subtitle            string       `json:"subtitle"`
				Authors             []string     `json:"authors"`
				PublishedDate       string       `json:"publishedDate"`
				Description         string       `json:"description"`
				IndustryIdentifiers []Identifier `json:"industryIdentifiers"`
				Categories          []string     `json:"categories"`
				ImageLinks          ImageLinks   `json:"imageLinks"`
			} `json:"volumeInfo"`
		}
		if err := json.Unmarshal(respBody, &aux); err != nil {
			return 0, fmt.Errorf("%v: %w", errorCaller, err)
		}
		v := aux.VolumeInfo
		b := model.Book{
			ID:          uuid.Nil,
			Title:       v.Title,
			Subtitle:    v.Subtitle,
			Description: v.Description,
			Published:   parsePublishedDate(v.PublishedDate),
			ISBNs:       extractISBN(v.IndustryIdentifiers),
			CoverImage:  uuid.Nil,
			ThumbImage:  uuid.Nil,
		}
		if _, exists, err := s.book.ExistsByISBN(ctx, b.ISBNs...); err != nil {
			if !errors.Is(err, repository.ErrorNotFound) {
				return 0, fmt.Errorf("%v: %w", errorCaller, err)
			}
		} else if exists {
			return 0, nil
		}

		// Now we know the book does not exist, so we can store it
		// Store thumbnail image
		storeBlobbedUrl := func(url string, to *uuid.UUID) error {
			b, err := urlToBlob(ctx, url)
			if err != nil {
				return fmt.Errorf("%v: %w", errorCaller, err)
			}
			if err := s.blob.Create(ctx, b); err != nil {
				return fmt.Errorf("%v: %w", errorCaller, err)
			}
			*to = b.ID
			return nil
		}
		// Set thumbnail and cover images
		if v.ImageLinks.Thumbnail != "" {
			storeBlobbedUrl(v.ImageLinks.Thumbnail, &b.ThumbImage)
		} else if v.ImageLinks.SmallThumbnail != "" {
			storeBlobbedUrl(v.ImageLinks.SmallThumbnail, &b.ThumbImage)
		}

		if v.ImageLinks.ExtraLarge != "" {
			storeBlobbedUrl(v.ImageLinks.ExtraLarge, &b.CoverImage)
		} else if v.ImageLinks.Large != "" {
			storeBlobbedUrl(v.ImageLinks.Large, &b.CoverImage)
		} else if v.ImageLinks.Medium != "" {
			storeBlobbedUrl(v.ImageLinks.Medium, &b.CoverImage)
		} else if v.ImageLinks.Small != "" {
			storeBlobbedUrl(v.ImageLinks.Small, &b.CoverImage)
		} else {
			b.CoverImage = b.ThumbImage
		}

		// Set book ID
		id, err := uuid.NewV7()
		if err != nil {
			return 0, fmt.Errorf("%v: %w", errorCaller, err)
		}
		b.ID = id

		// Authors
		for _, authorName := range v.Authors {
			// Check if the author already exists
			author, exists, err := s.athr.ExistsByName(ctx, "")
			if err != nil {
				fmt.Println("TODO: handle error")
				/*if !errors.Is(err, repository.ErrorNotFound) {
					return 0, fmt.Errorf("%v: %w", errorCaller, err)
				}*/
			}
			if !exists {
				// Create a new author if it does not exist
				// TODO: find a method to get the author from a tertiary source
				// as google books does not provide a method of discriminating
				// between authors with the same name.
				author = &model.Author{
					ID:         uuid.New(),
					GivenName:  "",
					FamilyName: authorName,
				}
				if err := s.athr.Create(ctx, author); err != nil {
					return 0, fmt.Errorf("%v: %w", errorCaller, err)
				}
			}
			b.AuthorIDs = append(b.AuthorIDs, author.ID)
		}

		// Commit the book to the datastore
		if err := s.book.Create(ctx, &b); err != nil {
			return 0, fmt.Errorf("%v: %w", errorCaller, err)
		}
		return 1, nil
	}

	total := 0
	for _, link := range links {
		n, err := newBook(link)
		if err != nil {
			eCh <- err
			return
		}
		total += n
	}
	iCh <- total
}

func (s *BookScraper) Scrape(ctx context.Context, offset, limit int, query string) (int, error) {
	const errorCaller string = "scrape"
	// Encode the query to be URL-safe if it is not already
	if decoded, err := url.QueryUnescape(query); err == nil && decoded == query {
		query = url.QueryEscape(query)
	}

	rem := limit % maxLimit
	nchunks := limit / maxLimit
	agents := nchunks
	if rem != 0 {
		agents++
	}
	// asynchronous function to fetch book data from self-link
	// and store it (and its blob, author, etc) in the database
	// if it does not already exist
	var wg sync.WaitGroup
	wg.Add(agents)
	errCh := make(chan error, agents)
	addCh := make(chan int, agents)

	if rem != 0 {
		go s.scrapeGoogleBooks(ctx, offset+nchunks*maxLimit, rem, query, addCh, errCh, &wg)
	}
	for i := range nchunks {
		go s.scrapeGoogleBooks(ctx, offset+i*maxLimit, maxLimit, query, addCh, errCh, &wg)
	}
	wg.Wait()
	close(errCh)
	close(addCh)

	// Check for errors
	for err := range errCh {
		if err != nil {
			return 0, fmt.Errorf("%v: %w", errorCaller, err)
		}
	}
	total := 0
	for added := range addCh {
		total += added
	}
	// Return the number of books stored
	return total, nil
}

// Gets the book data using ISBN and converts it
func (s *BookScraper) ScrapeISBN(ctx context.Context, isbn model.ISBN) (int, error) {
	return s.Scrape(ctx, 0, 1, fmt.Sprintf("isbn:%s", isbn.String()))
}
