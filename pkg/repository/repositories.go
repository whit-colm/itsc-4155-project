package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/models"
)

// The repository is the unifying "thing" that details all the
// interactions with the underlying model
type Repository struct {
	Store StoreManager
	Book  BookManager
}

type StoreManager interface {
	// A ping verifies the connection to the datastore is still
	// available, returning an error if something doesn't work
	//
	// Use this for healthcheck endpoints
	Ping(ctx context.Context) error
	// Disconnect from the datastore.
	Disconnect() error
}

type BookManager interface {
	Create(ctx context.Context, book *models.Book) error
	Delete(ctx context.Context, book *models.Book) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Book, error)
	Search(ctx context.Context) ([]models.Book, error)
	GetByISBN(ctx context.Context, isbn models.ISBN) (uuid.UUID, *models.Book, error)
}
