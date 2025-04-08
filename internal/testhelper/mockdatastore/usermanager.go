package mockdatastore

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

// UserRepo implements UserRepo.
type UserRepo struct {
	mu         sync.RWMutex
	users      map[uuid.UUID]*model.User
	byGithubID map[string]*model.User
	byUsername map[model.Username]*model.User
}

var _ repository.UserManager = (*UserRepo)(nil)

func NewInMemoryUserManager() *UserRepo {
	return &UserRepo{
		users:      make(map[uuid.UUID]*model.User),
		byGithubID: make(map[string]*model.User),
		byUsername: make(map[model.Username]*model.User),
	}
}

func (m *UserRepo) cache(user *model.User) {
	m.byGithubID[user.GithubID] = user
	m.byUsername[user.Username] = user
}

func (m *UserRepo) Create(ctx context.Context, user *model.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	m.users[user.ID] = user
	m.cache(user)
	return nil
}

func (m *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[id]
	if !exists {
		return nil, repository.ErrorNotFound
	}
	return user, nil
}

func (m *UserRepo) Update(ctx context.Context, user *model.User) (*model.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	u, exists := m.users[user.ID]
	if !exists {
		return nil, repository.ErrorNotFound
	}

	m.users[u.ID] = user
	m.cache(user)
	return user, nil
}

func (m *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[id]
	if !exists {
		return repository.ErrorNotFound
	}

	delete(m.byGithubID, user.GithubID)
	delete(m.byUsername, user.Username)
	delete(m.users, id)
	return nil
}

func (m *UserRepo) ExistsByGithubID(ctx context.Context, ghid string) (bool, error) {
	// DO NOT TAKE THE MUTEX FOR THIS METHOD.

	if _, err := m.GetByGithubID(ctx, ghid); err != nil {
		return false, err
	}
	return true, nil
}

func (m *UserRepo) GetByGithubID(ctx context.Context, ghid string) (*model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.byGithubID[ghid]
	if !exists {
		return nil, repository.ErrorNotFound
	}
	return user, nil
}

func (m *UserRepo) GetByUsername(ctx context.Context, username model.Username) (*model.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.byUsername[username]
	if !exists {
		return nil, repository.ErrorNotFound
	}
	return user, nil
}

func (m *UserRepo) Permissions(ctx context.Context, userID uuid.UUID) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[userID]
	if !exists {
		return false, repository.ErrorNotFound
	}

	return user.Admin, nil
}
