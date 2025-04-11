package mockdatastore

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

// AuthorRepo implements AuthorManager.
type AuthorRepo[S comparable] struct {
	book    repository.BookManager[S]
	mut     sync.RWMutex
	authors map[uuid.UUID]*model.Author
}

var _ repository.AuthorManager[string] = (*AuthorRepo[string])(nil)

func NewInMemoryAuthorManager[S comparable]() *AuthorRepo[S] {
	return &AuthorRepo[S]{
		authors: make(map[uuid.UUID]*model.Author),
	}
}

// Book implements repository.AuthorManager.
func (m *AuthorRepo[S]) Book(ctx context.Context, bookID uuid.UUID) ([]*model.Author, error) {
	m.mut.RLock()
	defer m.mut.RUnlock()
	var result []*model.Author

	if b, err := m.book.GetByID(ctx, bookID); err != nil {
		return nil, err
	} else {
		result = make([]*model.Author, 0, len(b.AuthorIDs))
		for _, aID := range b.AuthorIDs {
			if a, exists := m.authors[aID]; !exists {
				return nil, repository.ErrorNotFound
			} else {
				result = append(result, a)
			}
		}
	}
	return result, nil
}

func (m *AuthorRepo[S]) Create(ctx context.Context, author *model.Author) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	if author.ID == uuid.Nil {
		author.ID = uuid.New()
	}
	m.authors[author.ID] = author
	return nil
}

func (m *AuthorRepo[S]) GetByID(ctx context.Context, id uuid.UUID) (*model.Author, error) {
	m.mut.RLock()
	defer m.mut.RUnlock()

	author, exists := m.authors[id]
	if !exists {
		return nil, fmt.Errorf("author not found")
	}
	return author, nil
}

// Delete implements repository.AuthorManager.
func (m *AuthorRepo[S]) Delete(ctx context.Context, authorID uuid.UUID) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	delete(m.authors, authorID)
	return nil
}

// Search implements repository.AuthorManager.
func (m *AuthorRepo[S]) Search(ctx context.Context, offset int, limit int, query ...string) ([]repository.SearchResult[model.Author], error) {
	panic("unimplemented")
}

// Update implements repository.AuthorManager.
func (m *AuthorRepo[S]) Update(ctx context.Context, to *model.Author) (*model.Author, error) {
	m.mut.Lock()
	defer m.mut.Unlock()

	m.authors[to.ID] = to
	return to, nil
}
