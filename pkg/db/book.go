package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whit-colm/itsc-4155-project/pkg/models"
)

type BookRepositoryManager interface {
	Create(ctx context.Context, book *models.Book) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Book, error)
	Search(ctx context.Context) ([]models.Book, error)
}

type bookRepository struct {
	db *pgxpool.Pool
}

func NewBookRepository(pool *pgxpool.Pool) BookRepositoryManager {
	return &bookRepository{db: pool}
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
