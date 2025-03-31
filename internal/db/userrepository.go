package db

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"

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

	handle, disc := t.Username.Components()

	if _, err = tx.Exec(ctx,
		`INSERT INTO users (id, github_id, display_name, handle,
		 	discriminator, email, avatar, superuser)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		t.ID, t.GithubID, t.DisplayName, handle, disc, t.Email,
		t.Avatar, t.Admin,
	); err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
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

	if _, err := tx.Exec(ctx,
		`DELETE FROM users u
		 WHERE u.id = $1`,
		id,
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

// GetByGithubID implements repository.UserManager.
func (u *userRepository) GetByGithubID(ctx context.Context, ghid string) (*model.User, error) {
	var user model.User

	var handle string
	var discriminator uint16
	if err := u.db.QueryRow(ctx,
		`SELECT 
			 u.id,
			 u.github_id,
			 u.display_name,
			 u.pronouns,
			 u.handle,
			 u.discriminator,
			 u.email,
			 u.avatar,
		 	 u.superuser
		 FROM users u
		 WHERE u.github_id = $1
		 GROUP BY u.id`,
		ghid,
	).Scan(&user.ID, &user.GithubID, &user.DisplayName, &user.Pronouns,
		&handle, &discriminator, &user.Email, &user.Avatar, &user.Admin,
	); err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	user.Username = model.Username{}
	return &user, nil
}

// GetByID implements repository.UserManager.
func (u *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var user model.User

	var handle string
	var discriminator int16
	if err := u.db.QueryRow(ctx,
		`SELECT 
			 u.id,
			 u.github_id,
			 u.display_name,
			 u.pronouns,
			 u.handle,
			 u.discriminator,
			 u.email,
			 u.avatar,
		 	 u.superuser
		 FROM users u
		 WHERE u.id = $1
		 GROUP BY u.id`,
		id,
	).Scan(&user.ID, &user.GithubID, &user.DisplayName, &user.Pronouns,
		&handle, &discriminator, &user.Email, &user.Avatar, &user.Admin,
	); err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	if uname, err := model.UsernameFromComponents(
		handle, int(discriminator),
	); err != nil {
		return nil, fmt.Errorf("generate username: %w", err)
	} else {
		user.Username = uname
	}
	return &user, nil
}

// GetByUserHandle implements repository.UserManager.
func (u *userRepository) GetByUserHandle(ctx context.Context, username string) (*model.User, error) {
	var user model.User

	var handle string
	var discriminator int16
	if err := u.db.QueryRow(ctx,
		`SELECT 
			 u.id,
			 u.github_id,
			 u.display_name,
			 u.pronouns,
			 u.handle,
			 u.discriminator,
			 u.email,
			 u.avatar,
		 	 u.superuser
		 FROM users u
		 WHERE (u.handle || '#' || lpad(u.discriminator::TEXT, 4, '0')) = $1
		 GROUP BY u.id`,
		username,
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

// Search implements repository.UserManager.
func (u *userRepository) Search(ctx context.Context) ([]model.User, error) {
	panic("unimplemented")
}

// Update implements repository.UserManager.
func (u *userRepository) Update(ctx context.Context, t *model.User) (*model.User, error) {
	c, err := u.GetByID(ctx, t.ID)
	if err != nil {
		return nil, fmt.Errorf("current ")
	}
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	changes := generateDifferences(*c, *t)
	if _, found := changes["id"]; found {
		return nil, fmt.Errorf("users have different IDs")
	}
	// Handle special code for username
	// This is inherently bad code because if the column names change
	// we won't know.
	// TODO: AWFUL code. DO NOT USE
	if usernameAny, found := changes["username"]; found {
		if username, cast := usernameAny.(string); !cast {
			return nil, fmt.Errorf("username typecast: expect `string`, got `%v`",
				reflect.TypeOf(usernameAny))
		} else {
			s := strings.Split(username, "#")
			if len(s) != 2 {
				return nil,
					fmt.Errorf("username invalid constituent length: expect `2`, got `%d`",
						len(s))
			}
			if d, err := u.findDiscriminator(ctx, s[0]); err != nil {
				return nil, fmt.Errorf("generate new discriminator: %w", err)
			} else {
				changes["discriminator"] = d
			}
		}
		delete(changes, "username")
	}

	for k, v := range changes {
		if _, err = tx.Exec(ctx,
			`UPDATE users SET $1 = $2 WHERE id=$3`,
			k, v, t.ID,
		); err != nil {
			return nil, fmt.Errorf("update user: %w", err)
		}
	}

	return t, tx.Commit(ctx)
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

func generateDifferences[T any](t1, t2 T) map[string]any {
	changes := make(map[string]any)

	v1 := reflect.ValueOf(t1)
	v2 := reflect.ValueOf(t2)

	if v1.Kind() != reflect.Struct || v1.Type() != v2.Type() {
		return changes
	}

	t := v1.Type()

	for i := range v1.NumField() {
		f1 := v1.Field(i).Interface()
		f2 := v2.Field(i).Interface()

		if !reflect.DeepEqual(f1, f2) {
			tag := t.Field(i).Tag.Get("json")
			if tag == "" {
				return make(map[string]any)
			}
			changes[tag] = f2
		}
	}
	return changes
}
