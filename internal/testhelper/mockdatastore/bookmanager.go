package mockdatastore

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

// BookRepo implements BookRepo.
type BookRepo[S comparable] struct {
	athr       repository.AuthorManager[S]
	mut        sync.RWMutex
	books      map[uuid.UUID]*model.Book
	byISBN     map[model.ISBN]*model.Book
	byAuthorID map[uuid.UUID][]*model.Book
}

var _ repository.BookManager[string] = (*BookRepo[string])(nil)

func NewInMemoryBookManager[S comparable]() *BookRepo[S] {
	return &BookRepo[S]{
		books:      make(map[uuid.UUID]*model.Book),
		byISBN:     make(map[model.ISBN]*model.Book),
		byAuthorID: make(map[uuid.UUID][]*model.Book),
	}
}

func (m *BookRepo[S]) reindex() {
	// Delete old indexes
	m.byISBN = make(map[model.ISBN]*model.Book)
	m.byAuthorID = make(map[uuid.UUID][]*model.Book)
	for _, v := range m.books {
		for _, isbn := range v.ISBNs {
			m.byISBN[isbn] = v
		}
		for _, aID := range v.AuthorIDs {
			m.byAuthorID[aID] = append(m.byAuthorID[aID], v)
		}
	}
}

func (m *BookRepo[S]) Create(ctx context.Context, book *model.Book) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	if book.ID == uuid.Nil {
		book.ID = uuid.New()
	}

	m.books[book.ID] = book
	m.reindex()
	return nil
}

func (m *BookRepo[S]) GetByID(ctx context.Context, id uuid.UUID) (*model.Book, error) {
	m.mut.RLock()
	defer m.mut.RUnlock()

	if book, exists := m.books[id]; !exists {
		return nil, repository.ErrNotFound
	} else {
		return book, nil
	}
}

func (m *BookRepo[S]) Update(ctx context.Context, book *model.Book) (*model.Book, error) {
	m.mut.Lock()
	defer m.mut.Unlock()

	if c, exists := m.books[book.ID]; !exists {
		return nil, repository.ErrNotFound
	} else {
		m.books[book.ID] = c
	}
	m.reindex()
	return m.books[book.ID], nil
}

func (m *BookRepo[S]) Delete(ctx context.Context, id uuid.UUID) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	if book, exists := m.books[id]; !exists {
		return repository.ErrNotFound
	} else {
		delete(m.books, book.ID)
	}
	m.reindex()
	return nil
}

func (m *BookRepo[S]) GetByISBN(ctx context.Context, isbn model.ISBN) (*model.Book, error) {
	m.mut.RLock()
	defer m.mut.RUnlock()

	if book, exists := m.byISBN[isbn]; !exists {
		return nil, repository.ErrNotFound
	} else {
		return book, nil
	}
}

func (m *BookRepo[S]) ExistsByISBN(ctx context.Context, isbns ...model.ISBN) (*model.Book, bool, error) {
	m.mut.RLock()
	defer m.mut.RUnlock()
	for _, isbn := range isbns {
		if book, exists := m.byISBN[isbn]; exists {
			return book, true, nil
		}
	}
	return nil, false, nil
}

func (m *BookRepo[S]) Author(ctx context.Context, authorID uuid.UUID) ([]*model.Book, error) {
	m.mut.RLock()
	defer m.mut.RUnlock()

	if books, exists := m.byAuthorID[authorID]; !exists {
		return nil, repository.ErrNotFound
	} else {
		return books, nil
	}
}

// Search implements repository.BookManager.
func (m *BookRepo[S]) Search(ctx context.Context, offset int, limit int, query ...string) ([]repository.SearchResult[model.BookSummary], []repository.AnyScoreItemer, error) {
	panic("unimplemented")
}

// Summarize implements repository.BookManager.
func (m *BookRepo[S]) Summarize(context.Context, *model.Book) (*model.BookSummary, error) {
	panic("unimplemented")
}
