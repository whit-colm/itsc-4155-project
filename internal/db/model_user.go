package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type userRepository struct {
	db *pgxpool.Pool
}

// Useful to check that a type implements an interface
var _ repository.UserManager = (*userRepository)(nil)

func newUserRepository(psql *postgres) repository.UserManager {
	return &userRepository{db: psql.db}
}

// Create implements repository.UserManager.
func (u *userRepository) Create(ctx context.Context, t *model.User) error {
	panic("unimplemented")
}

// Delete implements repository.UserManager.
func (u *userRepository) Delete(ctx context.Context, t *model.User) error {
	panic("unimplemented")
}

// ExistsByGithubID implements repository.UserManager.
func (u *userRepository) ExistsByGithubID(ctx context.Context, ghid string) (bool, error) {
	panic("unimplemented")
}

// GetByGithubID implements repository.UserManager.
func (u *userRepository) GetByGithubID(ctx context.Context, ghid string) (*model.User, error) {
	panic("unimplemented")
}

// GetByID implements repository.UserManager.
func (u *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	panic("unimplemented")
}

// GetByUserHandle implements repository.UserManager.
func (u *userRepository) GetByUserHandle(ctx context.Context, handle string) (*model.User, error) {
	panic("unimplemented")
}

// Search implements repository.UserManager.
func (u *userRepository) Search(ctx context.Context) ([]model.User, error) {
	panic("unimplemented")
}

// Update implements repository.UserManager.
func (u *userRepository) Update(ctx context.Context, t *model.User) (*model.User, error) {
	panic("unimplemented")
}
