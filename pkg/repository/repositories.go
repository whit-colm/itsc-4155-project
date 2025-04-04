package repository

import (
	"context"
	"crypto"
	"time"

	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
)

// The repository is the unifying "thing" that details all the
// interactions with the underlying model
//
// TODO: I don't like this.
type Repository struct {
	Store   StoreManager
	Auth    AuthManager
	User    UserManager
	Author  AuthorManager
	Book    BookManager
	Blob    BlobManager
	Comment CommentManager
}

type StoreManager interface {
	Connect(ctx context.Context, args ...any) error
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
	Create(context.Context, *T) error
	GetByID(context.Context, uuid.UUID) (*T, error)
	Update(context.Context, uuid.UUID, *T) (*T, error)
	Delete(context.Context, uuid.UUID) error
}

type AuthManager interface {
	KeyPair(ctx context.Context) (crypto.PublicKey, crypto.Signer, error)
	Public(ctx context.Context) (crypto.PublicKey, error)
	Expiry(ctx context.Context) (time.Time, error)
	Rotate(ctx context.Context, ttl time.Duration) (crypto.Signer, error)
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
	GetByISBN(ctx context.Context, isbn model.ISBN) (*model.Book, error)
	Authors(ctx context.Context, bookID uuid.UUID) ([]*model.Author, error)
}

type BookSummaryManager interface {
	Searcher[model.BookSummary]
	Summarize(ctx context.Context, book *model.Book) (*model.BookSummary, error)
}

type AuthorManager interface {
	CRUDmanager[model.Author]
	Searcher[model.Author]
	Books(ctx context.Context, authorID uuid.UUID) ([]*model.Book, error)
}

type UserManager interface {
	CRUDmanager[model.User]
	Searcher[model.User]
	Permissions(ctx context.Context, userID uuid.UUID) (bool, error)
	GetByGithubID(ctx context.Context, ghid string) (*model.User, error)
	ExistsByGithubID(ctx context.Context, ghid string) (bool, error)
	GetByUserHandle(ctx context.Context, handle string) (*model.User, error)
	// Gets all comments the user has voted on as a yucky tuple
	//  - CommentID is the UUID of the comment
	//  - Vote is 1 for positive, -1 for negative. No 0's
	VotedComments(ctx context.Context, userID uuid.UUID) ([]*struct {
		CommentID uuid.UUID
		Vote      int
	}, error)
}

type CommentManager interface {
	CRUDmanager[model.Comment]
	GetBookComments(ctx context.Context, bookID uuid.UUID) ([]*model.Comment, error)
	GetAuthor(ctx context.Context, commentID uuid.UUID) (uuid.UUID, error)
	// Vote on a comment, a vote only ever counts for 1, therefore:
	//  - vote > 0 -> adds 1 to the total
	//  - vote < 0 -> removes 1 from the total
	//  - vote = 0 -> removes any vote
	// The total votes are returned at the end
	Vote(ctx context.Context, commentID uuid.UUID, vote int) (int, error)
	// Gets if you have voted on a comment. Like Vote(), returns an int
	// where:
	//  -  1 -> Positive vote
	//	- -1 -> Negative vote
	//  -  0 -> No vote
	Voted(ctx context.Context, commentID uuid.UUID) (int, error)
}
