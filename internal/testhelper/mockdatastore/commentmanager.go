package mockdatastore

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

// CommentRepo implements CommentManager.
type CommentRepo[S comparable] struct {
	repo     *InMemoryRepository[S]
	mut      sync.RWMutex
	comments map[uuid.UUID]*model.Comment
}

var _ repository.CommentManager[string] = (*CommentRepo[string])(nil)

func NewInMemoryCommentManager[S comparable]() *CommentRepo[S] {
	return &CommentRepo[S]{
		comments: make(map[uuid.UUID]*model.Comment),
	}
}

// BookComments implements repository.CommentManager.
func (r *CommentRepo[S]) BookComments(ctx context.Context, bookID uuid.UUID) ([]*model.Comment, error) {
	r.mut.RLock()
	defer r.mut.RUnlock()
	var results []*model.Comment

	if _, err := r.repo.Book.GetByID(ctx, bookID); err != nil {
		return nil, err
	}
	for _, c := range r.comments {
		if c.Book == bookID {
			results = append(results, c)
		}
	}

	return results, nil
}

// Create implements repository.CommentManager.
func (r *CommentRepo[S]) Create(context.Context, *model.Comment) error {
	r.mut.Lock()
	defer r.mut.Unlock()

	panic("unimplemented")
}

// Delete implements repository.CommentManager.
func (r *CommentRepo[S]) Delete(context.Context, uuid.UUID) error {
	r.mut.Lock()
	defer r.mut.Unlock()

	panic("unimplemented")
}

// GetByID implements repository.CommentManager.
func (r *CommentRepo[S]) GetByID(context.Context, uuid.UUID) (*model.Comment, error) {
	r.mut.RLock()
	defer r.mut.RUnlock()

	panic("unimplemented")
}

// Search implements repository.CommentManager.
func (r *CommentRepo[S]) Search(ctx context.Context, offset int, limit int, query ...string) ([]repository.SearchResult[model.Comment], error) {
	r.mut.RLock()
	defer r.mut.RUnlock()

	panic("unimplemented")
}

// Update implements repository.CommentManager.
func (r *CommentRepo[S]) Update(context.Context, *model.Comment) (*model.Comment, error) {
	r.mut.Lock()
	defer r.mut.Unlock()

	panic("unimplemented")
}
