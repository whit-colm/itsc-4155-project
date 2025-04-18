package db

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	handle, discriminator := t.Username.Components()
	// Generate valid discriminator if using a zero-value
	if discriminator == 0 && !slices.Contains(model.ReservedHandles, handle) {
		discriminator, err = u.findDiscriminator(ctx, handle)
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}
	}

	if _, err = tx.Exec(ctx,
		`INSERT INTO users (id, github_id, display_name, handle,
		 	discriminator, email, avatar, superuser)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		t.ID, t.GithubID, t.DisplayName, handle, discriminator,
		t.Email, t.Avatar, t.Admin,
	); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	return tx.Commit(ctx)
}

// Delete implements repository.UserManager.
func (u *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var avatarID uuid.UUID
	// we use u.db rather than tx.db because if this fails
 	// we don't want to kill the transaction
 	// we do not care if this fails
 	u.db.QueryRow(ctx,
		`SELECT avatar FROM users
		 WHERE id = $1`,
		id,
	).Scan(&avatarID)

	if _, err := tx.Exec(ctx,
		`DELETE FROM users
		 WHERE id = $1`,
		id,
	); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	// Don't try to delete user avatar if they don't have one
	if avatarID == uuid.Nil {
		return tx.Commit(ctx)
	}

	if _, err := tx.Exec(ctx,
		`DELETE FROM blobs b
		 WHERE b.id = $1`,
		avatarID,
	); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	return tx.Commit(ctx)
}

// ExistsByGithubID implements repository.UserManager.
func (u *userRepository) ExistsByGithubID(ctx context.Context, ghid string) (exists bool, err error) {
	if e := u.db.QueryRow(ctx,
		`SELECT EXISTS (
		 	 SELECT 1 FROM users WHERE github_id = $1
		 )`,
		ghid,
	).Scan(&exists); e != nil && !errors.Is(e, pgx.ErrNoRows) {
		return false, fmt.Errorf("search user by github id: %w", e)
	}
	return
}

func (u *userRepository) getByColumn(ctx context.Context, col, match string) (*model.User, error) {
	var user model.User
	var handle string
	var discriminator int16

	// This is **STUPIDLY** dangerous, and I only use it here like I do
	// because it's not used externally.
	query := fmt.Sprintf(`SELECT 
			 u.id,
			 COALESCE(u.github_id, ''),
			 COALESCE(u.display_name, u.handle),
			 COALESCE(u.pronouns, ''),
			 u.handle,
			 u.discriminator,
			 COALESCE(u.email, ''),
			 u.avatar,
		 	 u.superuser
		 FROM users u
		 WHERE %v = $1
		 GROUP BY u.id`,
		col,
	)
	if err := u.db.QueryRow(ctx,
		query, match,
	).Scan(&user.ID, &user.GithubID, &user.DisplayName, &user.Pronouns,
		&handle, &discriminator, &user.Email, &user.Avatar, &user.Admin,
	); err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	if uname, err := model.UsernameFromComponents(handle, discriminator); err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	} else {
		user.Username = uname
	}
	return &user, nil
}

// GetByGithubID implements repository.UserManager.
func (u *userRepository) GetByGithubID(ctx context.Context, ghid string) (*model.User, error) {
	return u.getByColumn(ctx, "u.github_id", ghid)
}

// GetByID implements repository.UserManager.
func (u *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	return u.getByColumn(ctx, "u.id", id.String())
}

// GetByUsername implements repository.UserManager.
func (u *userRepository) GetByUsername(ctx context.Context, username model.Username) (*model.User, error) {
	return u.getByColumn(ctx, "(u.handle || '#' || lpad(u.discriminator::TEXT, 4, '0'))", username.String())
}

func (u *userRepository) Permissions(ctx context.Context, userID uuid.UUID) (bool, error) {
	const errorCaller string = "get user permissions"
	var perms bool

	if err := u.db.QueryRow(ctx,
		`SELECT superuser FROM users WHERE id = $1`, userID,
	).Scan(&perms); err != nil {
		return false, fmt.Errorf("%v: %w", errorCaller, err)
	}
	return perms, nil
}

// Search implements repository.UserManager.
func (u *userRepository) Search(ctx context.Context) ([]model.User, error) {
	panic("unimplemented")
}

// Update implements repository.UserManager.
func (u *userRepository) Update(ctx context.Context, to *model.User) (*model.User, error) {
	const errorCaller string = "update user"
	from, err := u.GetByID(ctx, to.ID)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", errorCaller, err)
	}
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", errorCaller, err)
	}
	defer tx.Rollback(ctx)

	if from.ID != to.ID {
		return nil, fmt.Errorf("%v: users have different IDs", errorCaller)
	}

	handle, discriminator := to.Username.Components()
	if from.Username != to.Username {
		if v, err := u.validateNewUsername(ctx, to.Username.String()); err != nil || !v {
			discriminator, err = u.findDiscriminator(ctx, handle)
			if err != nil {
				return nil, fmt.Errorf("%v: %w", errorCaller, err)
			}
		}
	}

	if _, err = tx.Exec(ctx,
		`UPDATE users SET (
			 github_id,
			 display_name,
			 pronouns,
			 handle,
			 discriminator,
			 email,
			 avatar
		 ) = (
			 $2, $3, $4, $5, $6, $7, $8
		 ) WHERE id=$1`,
		to.ID, to.GithubID, to.DisplayName, to.Pronouns, handle,
		discriminator, to.Email, to.Avatar,
	); err != nil {
		return nil, fmt.Errorf("%v: %w", errorCaller, err)
	}

	return to, tx.Commit(ctx)
}

func (u *userRepository) validateNewUsername(ctx context.Context, username string) (bool, error) {
	var exists bool
	if err := u.db.QueryRow(ctx,
		`SELECT EXISTS (
			 SELECT 1 FROM users WHERE 
			 (handle || '#' || lpad(discriminator::TEXT, 4, '0')) = $1
		 )`, username,
	).Scan(&exists); err != nil {
		return false, fmt.Errorf("query: %w", err)
	}
	return exists, nil
}

func (u *userRepository) findDiscriminator(ctx context.Context, handle string) (int16, error) {
	var atCapacity bool
	if err := u.db.QueryRow(ctx,
		`SELECT COUNT(1) >= 9999 FROM users WHERE handle = $1`, handle,
	).Scan(&atCapacity); err != nil {
		return -1, fmt.Errorf("check capacity: %w", err)
	} else if atCapacity {
		return -1, fmt.Errorf("maximum number of usernames with this handle reached")
	}

	var discriminator int16
	if err := u.db.QueryRow(ctx,
		`SELECT d FROM generate_series(1,9999) AS d
		 WHERE NOT EXISTS (
		 	 SELECT 1
			 FROM users
			 WHERE handle = $1 AND discriminator = d
		 )
		 ORDER BY random()
		 LIMIT 1`, handle,
	).Scan(&discriminator); err != nil {
		return -1, fmt.Errorf("generate discriminator: %w", err)
	}
	return discriminator, nil
}
