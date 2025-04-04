package model

import (
	"cloud.google.com/go/civil"
	"github.com/google/uuid"
)

// An Books
type Book struct {
	ID        uuid.UUID  `json:"id"`
	ISBNs     []ISBN     `json:"isbns"`
	Title     string     `json:"title"`
	AuthorIDs uuid.UUIDs `json:"authors"`
	Published civil.Date `json:"published"`
}
