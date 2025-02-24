package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/models"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type bookRepository struct {
	db *postgres
}

func newBookRepository(psql *postgres) repository.BookManager {
	return &bookRepository{db: psql}
}

// Create implements BookRepositoryManager.
func (b *bookRepository) Create(ctx context.Context, book *models.Book) error {
	panic("unimplemented")
}

// GetByID implements BookRepositoryManager.
func (b *bookRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Book, error) {
	panic("unimplemented")
}

// Search implements BookRepositoryManager.
func (b *bookRepository) Search(ctx context.Context) ([]models.Book, error) {
	panic("unimplemented")
}
