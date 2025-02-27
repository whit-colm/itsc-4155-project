package repository

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"cloud.google.com/go/civil"
	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
)

var br TestingBookRepository

func init() {
	var eb []model.Book = []model.Book{
		{
			ID: uuid.MustParse("0124e053-3580-7000-8794-db4a97089840"),
			ISBNs: []model.ISBN{
				model.MustNewISBN("0141439602", model.ISBN10),
				model.MustNewISBN("9780141439600", model.ISBN13),
			},
			Title:     "A Tale of Two Cities",
			Author:    "Charles Dickens",
			Published: civil.Date{Year: 1859, Month: time.November, Day: 26},
		}, {
			ID: uuid.MustParse("0124e053-3580-7000-875a-c17e9ba5023c"),
			ISBNs: []model.ISBN{
				model.MustNewISBN("0156012197", model.ISBN10),
				model.MustNewISBN("9780156012195", model.ISBN13),
			},
			Title:     "The Little Prince",
			Author:    "Antoine de Saint-Exupéry",
			Published: civil.Date{Year: 1943, Month: time.April},
		}, {
			ID: uuid.MustParse("0124e053-3580-7000-9127-dd33bb29c893"),
			ISBNs: []model.ISBN{
				model.MustNewISBN("0062315005", model.ISBN10),
				model.MustNewISBN("9780061122415", model.ISBN13),
			},
			Title:     "The Alchemist",
			Author:    "Paulo Coelho",
			Published: civil.Date{Year: 1988},
		},
	}

	for _, v := range eb {
		j, _ := json.Marshal(v)
		br.books.Store(v.ID, j)

		for _, w := range v.ISBNs {
			br.isbns.Store(w, v.ID)
		}
	}
}

var testBook model.Book = model.Book{
	ID: uuid.MustParse("0124e053-3580-7000-a59a-fb9e45afdc80"),
	ISBNs: []model.ISBN{
		model.MustNewISBN("0062073486", model.ISBN10),
		model.MustNewISBN("978-0062073488", model.ISBN13),
	},
	Title:     "And Then There Were None",
	Author:    "Agatha Christie",
	Published: civil.Date{Year: 1939, Month: time.November, Day: 6},
}

/// Actual Test Suite ///

func TestCreate(t *testing.T) {
	if err := br.Create(t.Context(), &testBook); err != nil {
		t.Errorf("could not create book `%v`: %s", testBook, err)
	}
}

func TestGetByID(t *testing.T) {
	if b, err := br.GetByID(t.Context(), testBook.ID); err != nil {
		t.Errorf("error finding known UUID: %s", err)
	} else if b.ID != testBook.ID || /* Yes this is evil, I miss Rust */
		b.Author != testBook.Author ||
		b.Title != testBook.Title ||
		b.Published != testBook.Published ||
		// TODO: This is an actually unwell way to do it. O(N^2).
		func(s1, s2 []model.ISBN) bool {
			if len(s1) != len(s2) {
				return false
			}
			for _, v1 := range s1 {
				for _, v2 := range s2 {
					if !reflect.DeepEqual(v1, v2) {
						return false
					}
				}
			}
			return true
		}(b.ISBNs, testBook.ISBNs) {
		t.Errorf("inequality between fetched and known book: want %v; have %v", *b, testBook)
	}

	deadUUID, err := uuid.NewV7()
	if err != nil {
		t.Errorf("Error generating dead UUID: %s", err)
	}
	if b, _ := br.GetByID(t.Context(), deadUUID); b != nil {
		t.Errorf("unexpected found book for dead UUID: %v", b)
	}
}

func TestGetByISBN(t *testing.T) {
	if _, b, err := br.GetByISBN(t.Context(), testBook.ISBNs[0]); err != nil {
		t.Errorf("error finding known ISBN: %s", err)
	} else if b.ID != testBook.ID || /* Yes this is evil, I miss Rust */
		b.Author != testBook.Author ||
		b.Title != testBook.Title ||
		b.Published != testBook.Published ||
		// TODO: This is an actually unwell way to do it. O(N^2).
		func(s1, s2 []model.ISBN) bool {
			if len(s1) != len(s2) {
				return false
			}
			for _, v1 := range s1 {
				for _, v2 := range s2 {
					if !reflect.DeepEqual(v1, v2) {
						return false
					}
				}
			}
			return true
		}(b.ISBNs, testBook.ISBNs) {
		t.Errorf("inequality between fetched and known book: want %v; have %v", *b, testBook)
	}

	deadISBN := model.MustNewISBN("978-1408855652")
	if i, b, _ := br.GetByISBN(t.Context(), deadISBN); i != uuid.Nil || b != nil {
		t.Errorf("unexpected found book for dead ISBN: %v", i)
	}
}

func TestDelete(t *testing.T) {
	if err := br.Delete(t.Context(), &testBook); err != nil {
		t.Errorf("could not delete book `%v`: %s", testBook, err)
	}
}
