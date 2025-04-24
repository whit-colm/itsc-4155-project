package model

import (
	"cloud.google.com/go/civil"
	"github.com/google/uuid"
)

const BookSummaryApiVersion string = "booksummary.itsc-4155-group-project.edu.whits.io/v1alpha2"

type BookSummary struct {
	ID          uuid.UUID  `json:"id"`
	ISBNs       []ISBN     `json:"isbns"`
	Title       string     `json:"title"`
	Subtitle    string     `json:"subtitle,omitempty"`
	Description string     `json:"description"`
	Authors     []Author   `json:"authors"`
	Published   civil.Date `json:"published"`
	ThumbImage  uuid.UUID  `json:"bref_thumbnail_image,omitempty"`
}

func (b BookSummary) APIVersion() string {
	return BookSummaryApiVersion
}
