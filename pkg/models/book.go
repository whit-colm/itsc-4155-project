package models

import (
	"cloud.google.com/go/civil"
	"github.com/google/uuid"
)

// An Books
type Book struct {
	ID        uuid.UUID       `json:"id"`
	ISBN      []ISBNFormatter `json:"isbns"`
	Title     string          `json:"title"`
	Author    string          `json:"author"`
	Published civil.Date      `json:"published"`
}

func (b *Book) UnmarshalJSON(data []byte) error {
	return nil
}
