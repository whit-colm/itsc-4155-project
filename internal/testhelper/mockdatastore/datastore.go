package testdatastore

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
)

// InMemoryRepository implements the repository interfaces using in-memory data structures.
type InMemoryRepository struct {
	Store   *InMemoryStoreManager
	Auth    *InMemoryAuthManager
	User    *InMemoryUserManager
	Author  *InMemoryAuthorManager
	Book    *InMemoryBookManager
	Blob    *InMemoryBlobManager
	Comment *InMemoryCommentManager
}

// NewInMemoryRepository creates a new repository with all in-memory managers.
func NewInMemoryRepository() *InMemoryRepository {
	repo := &InMemoryRepository{
		Store:   &InMemoryStoreManager{},
		Auth:    &InMemoryAuthManager{},
		User:    NewInMemoryUserManager(),
		Author:  NewInMemoryAuthorManager(),
		Book:    NewInMemoryBookManager(),
		Blob:    NewInMemoryBlobManager(),
		Comment: NewInMemoryCommentManager(),
	}

	// Link child managers back to the repository for cross-manager access
	repo.User.repo = repo
	repo.Author.repo = repo
	repo.Book.repo = repo
	repo.Comment.repo = repo

	return repo
}

// InMemoryStoreManager implements StoreManager.
type InMemoryStoreManager struct{}

func (m *InMemoryStoreManager) Connect(ctx context.Context, args ...any) error { return nil }
func (m *InMemoryStoreManager) Ping(ctx context.Context) error                 { return nil }
func (m *InMemoryStoreManager) Disconnect() error                              { return nil }

// InMemoryAuthManager implements AuthManager.
type InMemoryAuthManager struct {
	mu     sync.RWMutex
	pubKey crypto.PublicKey
	signer crypto.Signer
	expiry time.Time
}

func (m *InMemoryAuthManager) KeyPair(ctx context.Context) (crypto.PublicKey, crypto.Signer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pubKey, m.signer, nil
}

func (m *InMemoryAuthManager) Public(ctx context.Context) (crypto.PublicKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pubKey, nil
}

func (m *InMemoryAuthManager) Expiry(ctx context.Context) (time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.expiry, nil
}

func (m *InMemoryAuthManager) Rotate(ctx context.Context, ttl time.Duration) (crypto.Signer, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	m.signer = privateKey
	m.pubKey = privateKey.Public()
	m.expiry = time.Now().Add(ttl)

	return m.signer, nil
}

// InMemoryUserManager implements UserManager.
type InMemoryUserManager struct {
	repo       *InMemoryRepository
	mu         sync.RWMutex
	users      map[uuid.UUID]*model.User
	byGithubID map[string]*model.User
	byHandle   map[string]*model.User
}

func NewInMemoryUserManager() *InMemoryUserManager {
	return &InMemoryUserManager{
		users:      make(map[uuid.UUID]*model.User),
		byGithubID: make(map[string]*model.User),
		byHandle:   make(map[string]*model.User),
	}
}

func (m *InMemoryUserManager) Create(ctx context.Context, user *model.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	m.users[user.ID] = user
	m.byGithubID[user.GithubID] = user
	m.byHandle[user.Username.String()] = user
	return nil
}

func (m *InMemoryUserManager) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[id]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (m *InMemoryUserManager) Update(ctx context.Context, id uuid.UUID, user *model.User) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.users[id]; !exists {
		return nil, fmt.Errorf("user not found")
	}

	// Update indexes
	delete(m.byGithubID, m.users[id].GithubID)
	delete(m.byHandle, m.users[id].Username.String())

	m.users[id] = user
	m.byGithubID[user.GithubID] = user
	m.byHandle[user.Username.String()] = user
	return user, nil
}

func (m *InMemoryUserManager) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[id]
	if !exists {
		return fmt.Errorf("user not found")
	}

	delete(m.byGithubID, user.GithubID)
	delete(m.byHandle, user.Username.String())
	delete(m.users, id)
	return nil
}

func (m *InMemoryUserManager) GetByGithubID(ctx context.Context, ghid string) (*model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.byGithubID[ghid]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (m *InMemoryUserManager) GetByUserHandle(ctx context.Context, handle string) (*model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.byHandle[handle]
	if !exists {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

// Other UserManager methods (Permissions, ExistsByGithubID, VotedComments) would follow similar patterns.

// InMemoryBookManager implements BookManager.
type InMemoryBookManager struct {
	repo   *InMemoryRepository
	mu     sync.RWMutex
	books  map[uuid.UUID]*model.Book
	byISBN map[model.ISBN]uuid.UUID
}

func NewInMemoryBookManager() *InMemoryBookManager {
	return &InMemoryBookManager{
		books:  make(map[uuid.UUID]*model.Book),
		byISBN: make(map[model.ISBN]uuid.UUID),
	}
}

func (m *InMemoryBookManager) Create(ctx context.Context, book *model.Book) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if book.ID == uuid.Nil {
		book.ID = uuid.New()
	}

	m.books[book.ID] = book
	for _, isbn := range book.ISBNs {
		m.byISBN[isbn] = book.ID
	}
	return nil
}

func (m *InMemoryBookManager) GetByID(ctx context.Context, id uuid.UUID) (*model.Book, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	book, exists := m.books[id]
	if !exists {
		return nil, fmt.Errorf("book not found")
	}
	return book, nil
}

func (m *InMemoryBookManager) GetByISBN(ctx context.Context, isbn model.ISBN) (*model.Book, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	bookID, exists := m.byISBN[isbn]
	if !exists {
		return nil, fmt.Errorf("book not found")
	}
	return m.GetByID(ctx, bookID)
}

func (m *InMemoryBookManager) Authors(ctx context.Context, bookID uuid.UUID) ([]*model.Author, error) {
	book, err := m.GetByID(ctx, bookID)
	if err != nil {
		return nil, err
	}

	var authors []*model.Author
	for _, authorID := range book.AuthorIDs {
		author, err := m.repo.Author.GetByID(ctx, authorID)
		if err != nil {
			return nil, err
		}
		authors = append(authors, author)
	}
	return authors, nil
}

// Other BookManager methods (Update, Delete, Search) would follow similar patterns.

// InMemoryAuthorManager implements AuthorManager.
type InMemoryAuthorManager struct {
	repo    *InMemoryRepository
	mu      sync.RWMutex
	authors map[uuid.UUID]*model.Author
}

func NewInMemoryAuthorManager() *InMemoryAuthorManager {
	return &InMemoryAuthorManager{
		authors: make(map[uuid.UUID]*model.Author),
	}
}

func (m *InMemoryAuthorManager) Create(ctx context.Context, author *model.Author) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if author.ID == uuid.Nil {
		author.ID = uuid.New()
	}
	m.authors[author.ID] = author
	return nil
}

func (m *InMemoryAuthorManager) GetByID(ctx context.Context, id uuid.UUID) (*model.Author, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	author, exists := m.authors[id]
	if !exists {
		return nil, fmt.Errorf("author not found")
	}
	return author, nil
}

// Other AuthorManager methods (Update, Delete, Search, Books) follow similar patterns.

// InMemoryBlobManager implements BlobManager.
type InMemoryBlobManager struct {
	mu    sync.RWMutex
	blobs map[uuid.UUID]*model.Blob
}

func NewInMemoryBlobManager() *InMemoryBlobManager {
	return &InMemoryBlobManager{
		blobs: make(map[uuid.UUID]*model.Blob),
	}
}

func (m *InMemoryBlobManager) Create(ctx context.Context, blob *model.Blob) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := io.ReadAll(blob.Content)
	if err != nil {
		return fmt.Errorf("read blob content: %w", err)
	}

	// Store content as a new bytes.Reader for each retrieval
	m.blobs[blob.ID] = &model.Blob{
		ID:       blob.ID,
		Metadata: blob.Metadata,
		Content:  bytes.NewReader(data),
	}
	return nil
}

func (m *InMemoryBlobManager) GetByID(ctx context.Context, id uuid.UUID) (*model.Blob, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	blob, exists := m.blobs[id]
	if !exists {
		return nil, fmt.Errorf("blob not found")
	}

	// Return a copy with a fresh reader
	data, _ := io.ReadAll(blob.Content)
	return &model.Blob{
		ID:       blob.ID,
		Metadata: blob.Metadata,
		Content:  bytes.NewReader(data),
	}, nil
}

// Update and Delete methods follow similar patterns.

// InMemoryCommentManager implements CommentManager.
type InMemoryCommentManager struct {
	repo     *InMemoryRepository
	mu       sync.RWMutex
	comments map[uuid.UUID]*model.Comment
	votes    map[uuid.UUID]map[uuid.UUID]int // commentID -> userID -> vote
}

func NewInMemoryCommentManager() *InMemoryCommentManager {
	return &InMemoryCommentManager{
		comments: make(map[uuid.UUID]*model.Comment),
		votes:    make(map[uuid.UUID]map[uuid.UUID]int),
	}
}

func (m *InMemoryCommentManager) Vote(ctx context.Context, commentID uuid.UUID, vote int) (int, error) {
	userID, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
		return 0, fmt.Errorf("userID not found in context")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	userVotes, exists := m.votes[commentID]
	if !exists {
		userVotes = make(map[uuid.UUID]int)
		m.votes[commentID] = userVotes
	}

	current := userVotes[userID]
	switch {
	case vote == 0:
		delete(userVotes, userID)
	case vote > 0:
		userVotes[userID] = 1
	case vote < 0:
		userVotes[userID] = -1
	}

	total := 0
	for _, v := range userVotes {
		total += v
	}
	return total, nil
}

// Other CommentManager methods (Create, GetByID, GetBookComments, etc.) follow similar patterns.
