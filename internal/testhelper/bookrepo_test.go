package testhelper

import (
	"encoding/json"
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
	} else if !IsBookEquals(*b, ExampleBook) {
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
	} else if !IsBookEquals(*b, ExampleBook) {
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
