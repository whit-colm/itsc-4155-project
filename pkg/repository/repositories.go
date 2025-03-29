package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
)

// The repository is the unifying "thing" that details all the
// interactions with the underlying model
//
// TODO: I don't like this.
type Repository struct {
	Store  StoreManager
	User   UserManager
	Author AuthorManager
	Book   BookManager
}

type Creator[T any] interface {
	New(t any) (T, error)
}

type StoreManager interface {
	Connect(args any) error
	// A ping verifies the connection to the datastore is still
	// available, returning an error if something doesn't work
	//
	// Use this for healthcheck endpoints
	Ping(ctx context.Context) error
	// Disconnect from the datastore.
	Disconnect() error
}

// The most fundamental manager type, which implements primitive CRUD
// operations on a generic model.
//
// Any more complicated types should instead have their own dedicated
// manager, which implements CRUDmanager.
type CRUDmanager[T any] interface {
	Create(ctx context.Context, t *T) error
	GetByID(ctx context.Context, id uuid.UUID) (*T, error)
	Update(ctx context.Context, t *T) (*T, error)
	Delete(ctx context.Context, t *T) error
}

// This is cooked for the time being. Do not use.
//
// In future there will be a change to the Search method adding a
// parameter for search terms
type Searcher[T any] interface {
	Search(ctx context.Context) ([]T, error)
}

type BlobManager interface {
	CRUDmanager[model.Blob]
}

type BookManager interface {
	CRUDmanager[model.Book]
	Searcher[model.Book]
	GetByISBN(ctx context.Context, isbn model.ISBN) (uuid.UUID, *model.Book, error)
}

type BookSummaryManager interface {
	Searcher[model.BookSummary]
	Summarize(ctx context.Context, book *model.Book) (*model.BookSummary, error)
}

type AuthorManager interface {
	CRUDmanager[model.Author]
	Searcher[model.Author]
	GetByBook(ctx context.Context, book model.Book) (*model.Author, error)
}

type UserManager interface {
	CRUDmanager[model.User]
	Searcher[model.User]
	GetByGithubID(ctx context.Context, ghid string) (*model.User, error)
	ExistsByGithubID(ctx context.Context, ghid string) (bool, error)
}
