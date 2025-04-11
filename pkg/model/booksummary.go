package model

import (
	"cloud.google.com/go/civil"
	"github.com/google/uuid"
)

type BookSummary struct {
	ID        uuid.UUID  `json:"id"`
	ISBNs     []ISBN     `json:"isbns"`
	Title     string     `json:"title"`
	Authors   []Author   `json:"authors"`
	Published civil.Date `json:"published"`
}
