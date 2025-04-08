package mockdatastore

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"sync"
	"time"
)

// InMemoryRepository implements the repository interfaces using in-memory data structures.
type InMemoryRepository[S comparable] struct {
	Store   *StoreRepo
	Auth    *AuthRepo
	User    *UserRepo
	Author  *AuthorRepo
	Book    *BookRepo
	Blob    *BlobRepo
	Comment *CommentRepo[S]
}

// NewInMemoryRepository creates a new repository with all in-memory managers.
func NewInMemoryRepository[S comparable]() *InMemoryRepository[S] {
	repo := &InMemoryRepository[S]{
		Store:   &StoreRepo{},
		Auth:    &AuthRepo{},
		User:    NewInMemoryUserManager(),
		Author:  NewInMemoryAuthorManager(),
		Book:    NewInMemoryBookManager(),
		Blob:    NewInMemoryBlobManager(),
		Comment: NewInMemoryCommentManager[S](),
	}

	// Link child managers back to the repository for cross-manager access
	repo.Author.book = repo.Book
	repo.Book.athr = repo.Author
	repo.Comment.repo = repo

	return repo
}

// StoreRepo implements StoreManager.
type StoreRepo struct{}

func (m *StoreRepo) Connect(ctx context.Context, args ...any) error { return nil }
func (m *StoreRepo) Ping(ctx context.Context) error                 { return nil }
func (m *StoreRepo) Disconnect() error                              { return nil }

// AuthRepo implements AuthManager.
type AuthRepo struct {
	mu     sync.RWMutex
	pubKey crypto.PublicKey
	signer crypto.Signer
	expiry time.Time
}

func (m *AuthRepo) KeyPair(ctx context.Context) (crypto.PublicKey, crypto.Signer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pubKey, m.signer, nil
}

func (m *AuthRepo) Public(ctx context.Context) (crypto.PublicKey, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.pubKey, nil
}

func (m *AuthRepo) Expiry(ctx context.Context) (time.Time, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.expiry, nil
}

func (m *AuthRepo) Rotate(ctx context.Context, ttl time.Duration) (crypto.Signer, error) {
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
