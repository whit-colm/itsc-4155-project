package models

import (
	"cloud.google.com/go/civil"
	"github.com/google/uuid"
)

// An Books
type Book struct {
	ID        uuid.UUID  `json:"id"`
	ISBNs     []ISBN     `json:"isbns"`
	Title     string     `json:"title"`
	Author    string     `json:"author"`
	Published civil.Date `json:"published"`
}

/*
func (b *Book) UnmarshalJSON(data []byte) error {
	var aux struct {
		ID        uuid.UUID         `json:"id"`
		ISBNs     []json.RawMessage `json:"isbns"`
		Title     string            `json:"title"`
		Author    string            `json:"author"`
		Published civil.Date        `json:"published"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	b.ID = aux.ID
	b.Title = aux.Title
	b.ISBNs = make([]ISBNFormatter, 0, len(aux.ISBNs))
	b.Author = aux.Author
	b.Published = aux.Published

	for _, raw := range aux.ISBNs {
		var jsonIsbn struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		}
		if err := json.Unmarshal(raw, &jsonIsbn); err != nil {
			return err
		}

		var isbn ISBNFormatter
		switch jsonIsbn.Type {
		case "isbn10":
			var i10 ISBN10
			if err := json.Unmarshal(raw, &i10); err != nil {
				return err
			}
			isbn = &i10
		case "isbn13":
			var i13 ISBN13
			if err := json.Unmarshal(raw, &i13); err != nil {
				return err
			}
			isbn = &i13
		}
		b.ISBNs = append(b.ISBNs, isbn)
	}
	return nil
}
*/
