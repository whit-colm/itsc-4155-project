package repository

import (
	"context"
	"crypto"
	"time"

	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
)

/********************/
/*** ABSTRACTIONS ***/
/********************/

// The repository is the unifying "thing" that details all the
// interactions with the underlying model
//
// TODO: I don't like this.
type Repository[S comparable] struct {
	Author  AuthorManager[S]
	Auth    AuthManager
	Blob    BlobManager
	Book    BookManager[S]
	Comment CommentManager[S]
	User    UserManager
	Store   StoreManager
	Vote    VoteManager
}

// The most fundamental manager type, which implements primitive CRUD
// operations on a generic model.
//
// Any more complicated types should instead have their own dedicated
// manager, which implements CRUDmanager.
//
// You *could* use CRUDmanager on its own but why would you?
type CRUDmanager[K comparable, T any] interface {
	Create(context.Context, *T) error
	GetByID(context.Context, K) (*T, error)
	// Return the updated object
	Update(context.Context, *T) (*T, error)
	Delete(context.Context, K) error
}

// Search repository against query parameters
//
// In addition to generic search query terms,
//
// This method returns a generic struct of a pointer to the item and a
// score. For multi-manager searches, use score as the merge value.
//
// TODO: query is string to make me not want to kill myself, should a
// real app use that generic S?
type Searcher[S comparable, T any] interface {
	// Searches the domain given a query
	//
	// The two non-error return values are effectively the same, except
	// for their types; use the AnyScoreItemer for JSON stuff and the
	// SearchResult[T] for typed internal searches
	Search(ctx context.Context, offset, limit int, query ...string) ([]SearchResult[T], []AnyScoreItemer, error)
}

type AnyScoreItemer interface {
	ItemAsAny() any
	ScoreValue() float64
}

type SearchResult[T any] struct {
	Item  *T
	Score float64
}

func (sr SearchResult[T]) ItemAsAny() any {
	return sr.Item
}

func (sr SearchResult[T]) ScoreValue() float64 {
	return sr.Score
}

/*******************************/
/*** TOP-LEVEL SYSTEM CONFIG ***/
/*******************************/

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

type AuthManager interface {
	KeyPair(ctx context.Context) (crypto.PublicKey, crypto.Signer, error)
	Public(ctx context.Context) (crypto.PublicKey, error)
	Expiry(ctx context.Context) (time.Time, error)
	Rotate(ctx context.Context, ttl time.Duration) (crypto.Signer, error)
}

type BlobManager interface {
	CRUDmanager[uuid.UUID, model.Blob]
}

/******************************/
/*** BOOKS, AUTHORS, GENRES ***/
/******************************/

type AuthorManager[S comparable] interface {
	CRUDmanager[uuid.UUID, model.Author]
	Searcher[S, model.Author]
	Book(ctx context.Context, bookID uuid.UUID) ([]*model.Author, error)
}

type BookManager[S comparable] interface {
	CRUDmanager[uuid.UUID, model.Book]
	Searcher[S, model.BookSummary]
	Summarize(context.Context, *model.Book) (*model.BookSummary, error)
	GetByISBN(context.Context, model.ISBN) (*model.Book, error)
	Author(ctx context.Context, authorID uuid.UUID) ([]*model.Book, error)
}

/*************************/
/*** USER INTERACTIONS ***/
/*************************/

type CommentManager[S comparable] interface {
	CRUDmanager[uuid.UUID, model.Comment]
	Searcher[S, model.Comment]
	BookComments(ctx context.Context, bookID uuid.UUID) ([]*model.Comment, error)
}

type UserManager interface {
	CRUDmanager[uuid.UUID, model.User]
	ExistsByGithubID(context.Context, string) (bool, error)
	GetByGithubID(context.Context, string) (*model.User, error)
	GetByUsername(context.Context, model.Username) (*model.User, error)
	Permissions(context.Context, uuid.UUID) (bool, error)
}

type VoteManager interface {
	// Gets all comments the user has voted on as a yucky tuple
	//  - CommentID is the UUID of the comment
	//  - Vote is 1 for positive, -1 for negative. No 0's
	UserVotes(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]int8, error)

	// Vote on a comment, a vote only ever counts for 1, therefore:
	//  - vote > 0 -> adds 1 to the total
	//  - vote < 0 -> removes 1 from the total
	//  - vote = 0 -> removes any vote
	// The total votes are returned at the end
	Vote(ctx context.Context, userID uuid.UUID, commentID uuid.UUID, vote int) (int, error)

	// Gets if a user has voted on some comments. Like Vote(), int
	// values returned follow the pattern:
	//  -  1 -> Positive vote
	//	- -1 -> Negative vote
	//  -  0 -> No vote
	Voted(ctx context.Context, userID uuid.UUID, commentIDs uuid.UUIDs) (map[uuid.UUID]int8, error)
}
