package repository

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
)

// TestingBookRepository is a useful tool when testing other packages
//
// Packages like endpoint should use this instead of the actual DB
// during testing.
type TestingBookRepository struct {
	// map[uuid.UUID][]byte
	books sync.Map

	// map[model.ISBN]uuid.UUID
	isbns sync.Map
}

var _ BookManager = (*TestingBookRepository)(nil)

// Create implements repository.BookManager.
func (t *TestingBookRepository) Create(ctx context.Context, book *model.Book) error {
	if v, exists := t.books.LoadOrStore(book.ID, *book); exists {
		return fmt.Errorf("book with ID %s already exists: '%v'", book.ID, v)
	}

	for i, v := range book.ISBNs {
		if w, exists := t.isbns.LoadOrStore(v, book.ID); exists {
			// So we can't really do transactions, so to reverse all of
			// this and make sure we don't end up with a polluted map
			// we have to do it ourselves.
			t.books.Delete(book.ID)

			for _, v := range book.ISBNs[:i] {
				t.isbns.Delete(v)
			}

			return fmt.Errorf("ISBN %s already exists.", w)
		}
	}
	return nil
}

// Delete implements repository.BookManager.
func (t *TestingBookRepository) Delete(ctx context.Context, book *model.Book) error {
	t.books.Delete(book.ID)
	for _, v := range book.ISBNs {
		t.isbns.Delete(v)
	}

	return nil
}

// GetByID implements repository.BookManager.
func (t *TestingBookRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Book, error) {
	v, found := t.books.Load(id)
	if !found {
		return nil, fmt.Errorf("could not find book matching id `%v`", id)
	}
	if b, ok := v.(model.Book); !ok {
		return nil, fmt.Errorf("could not cast to Book: %#v", v)
	} else {
		return &b, nil
	}
}

// GetByISBN implements repository.BookManager.
func (t *TestingBookRepository) GetByISBN(ctx context.Context, isbn model.ISBN) (uuid.UUID, *model.Book, error) {
	v, found := t.isbns.Load(isbn)
	if !found {
		return uuid.Nil, nil, fmt.Errorf("could not find book matching ISBN `%v`", isbn)
	}
	id, ok := v.(uuid.UUID)
	if !ok {
		return uuid.Nil, nil, fmt.Errorf("could not cast to UUID: %#v", v)
	}

	v, found = t.books.Load(id)
	if !found {
		return uuid.Nil, nil, fmt.Errorf("could not find book matching id `%v`", id)
	}
	b, ok := v.(model.Book)
	if !ok {
		return uuid.Nil, nil, fmt.Errorf("could not cast to Book: %#v", v)
	}
	return id, &b, nil
}

// Search implements repository.BookManager.
func (t *TestingBookRepository) Search(ctx context.Context) ([]model.Book, error) {
	panic("unimplemented")
}
