package model

import (
	"cloud.google.com/go/civil"
	"github.com/google/uuid"
)

const BookSummaryApiVersion string = "booksummary.itsc-4155-group-project.edu.whits.io/v1alpha1"

type BookSummary struct {
	ID        uuid.UUID  `json:"id"`
	ISBNs     []ISBN     `json:"isbns"`
	Title     string     `json:"title"`
	Authors   []Author   `json:"authors"`
	Published civil.Date `json:"published"`
}

func (b BookSummary) APIVersion() string {
	return BookSummaryApiVersion
}
