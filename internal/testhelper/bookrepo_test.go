package testhelper

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
)

var br TestingBookRepository

func init() {
	for _, v := range ExampleBooks {
		j, _ := json.Marshal(v)
		br.books.Store(v.ID, j)

		for _, w := range v.ISBNs {
			br.isbns.Store(w, v.ID)
		}
	}
}

func TestCreate(t *testing.T) {
	if err := br.Create(t.Context(), &ExampleBook); err != nil {
		t.Errorf("could not create book `%v`: %s", ExampleBook, err)
	}
}

func TestGetByID(t *testing.T) {
	if b, err := br.GetByID(t.Context(), ExampleBook.ID); err != nil {
		t.Errorf("error finding known UUID: %s", err)
	} else if b.ID != ExampleBook.ID || /* Yes this is evil, I miss Rust */
		b.Author != ExampleBook.Author ||
		b.Title != ExampleBook.Title ||
		b.Published != ExampleBook.Published ||
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
		}(b.ISBNs, ExampleBook.ISBNs) {
		t.Errorf("inequality between fetched and known book: want %v; have %v", *b, ExampleBook)
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
	if _, b, err := br.GetByISBN(t.Context(), ExampleBook.ISBNs[0]); err != nil {
		t.Errorf("error finding known ISBN: %s", err)
	} else if b.ID != ExampleBook.ID || /* Yes this is evil, I miss Rust */
		b.Author != ExampleBook.Author ||
		b.Title != ExampleBook.Title ||
		b.Published != ExampleBook.Published ||
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
		}(b.ISBNs, ExampleBook.ISBNs) {
		t.Errorf("inequality between fetched and known book: want %v; have %v", *b, ExampleBook)
	}

	deadISBN := model.MustNewISBN("978-1408855652")
	if i, b, _ := br.GetByISBN(t.Context(), deadISBN); i != uuid.Nil || b != nil {
		t.Errorf("unexpected found book for dead ISBN: %v", i)
	}
}

func TestDelete(t *testing.T) {
	if err := br.Delete(t.Context(), &ExampleBook); err != nil {
		t.Errorf("could not delete book `%v`: %s", ExampleBook, err)
	}
}
