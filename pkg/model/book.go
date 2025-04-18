package model

import (
	"cloud.google.com/go/civil"
	"github.com/google/uuid"
)

const BookApiVersion string = "book.itsc-4155-group-project.edu.whits.io/v1alpha1"

// An Books
type Book struct {
	ID        uuid.UUID  `json:"id"`
	ISBNs     []ISBN     `json:"isbns"`
	Title     string     `json:"title"`
	AuthorIDs uuid.UUIDs `json:"authors"`
	Published civil.Date `json:"published"`
}

func (b Book) APIVersion() string {
	return BookApiVersion
}
