package model

import (
	"cloud.google.com/go/civil"
	"github.com/google/uuid"
)

type BookSummary struct {
	ID               uuid.UUID  `json:"id"`
	ISBNs            []ISBN     `json:"isbns"`
	Title            string     `json:"title"`
	AuthorID         uuid.UUID  `json:"authorId"`
	AuthorFamilyName string     `json:"authorFamilyName"`
	AuthorGivenName  string     `json:"authorGivenName"`
	Published        civil.Date `json:"published"`
}
