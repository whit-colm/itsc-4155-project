package model

import (
	"cloud.google.com/go/civil"
	"github.com/google/uuid"
)

const BookApiVersion string = "book.itsc-4155-group-project.edu.whits.io/v1alpha2"

// An Books
type Book struct {
	ID          uuid.UUID  `json:"id"`
	ISBNs       []ISBN     `json:"isbns"`
	Title       string     `json:"title"`
	Subtitle    string     `json:"subtitle,omitempty"`
	Description string     `json:"description"`
	AuthorIDs   uuid.UUIDs `json:"authors"`
	Published   civil.Date `json:"published"`
	CoverImage  uuid.UUID  `json:"bref_cover_image,omitempty"`
	ThumbImage  uuid.UUID  `json:"bref_thumbnail_image,omitempty"`
}

func (b Book) APIVersion() string {
	return BookApiVersion
}
