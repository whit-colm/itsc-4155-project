package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type authorRepository struct {
	db *pgxpool.Pool
}

// Useful to check that a type implements an interface
var _ repository.AuthorManager = (*authorRepository)(nil)

func newAuthorRepository(psql *postgres) repository.AuthorManager {
	return &authorRepository{db: psql.db}
}

func (a *authorRepository) Books(ctx context.Context, bookID uuid.UUID) ([]*model.Book, error) {
	panic("unimplemented")
}

// Create implements repository.AuthorManager.
func (a *authorRepository) Create(ctx context.Context, author *model.Author) error {
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO authors (id, familyname, givenname)
		 VALUES ($1, $2, $3, $4)`,
		author.ID, author.GivenName, author.GivenName,
	)
	if err != nil {
		return fmt.Errorf("failed to insert author: %w", err)
	}

	return tx.Commit(ctx)
}

// Update implements repository.BookManager.
func (b *authorRepository) Update(ctx context.Context, to *model.Author) (*model.Author, error) {
	panic("unimplemented")
}

// Delete implements repository.AuthorManager.
func (a *authorRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`DELETE FROM authors a
		 WHERE a.id = $1`,
		id,
	); err != nil {
		return fmt.Errorf("delete author: %w", err)
	}

	return tx.Commit(ctx)
}

// GetByID implements repository.AuthorManager.
func (a *authorRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Author, error) {
	var author model.Author

	if err := a.db.QueryRow(ctx,
		`SELECT
			a.id,
			a.givenname,
			a.familyname
		FROM authors a
		WHERE a.id = $1
		GROUP BY a.id`,
		id,
	).Scan(
		&author.ID, &author.GivenName, &author.FamilyName,
	); err != nil {
		return nil, fmt.Errorf("could not retrieve author: %w", err)
	}
	return &author, nil
}

// Search implements repository.AuthorManager.
//
// Note that this is currently cooked, and just returns all authors in the db
// at a later time updates will be made to allow for search parameters.
func (a *authorRepository) Search(ctx context.Context) ([]model.Author, error) {
	authors := make([]model.Author, 0, 3)

	rows, err := a.db.Query(ctx,
		`SELECT
			a.id,
			a.givenname,
			a.familyname
		FROM authors a
		GROUP BY a.id`)

	if err != nil {
		return []model.Author{}, fmt.Errorf("unable to query db: %w", err)
	}

	for rows.Next() {
		var a model.Author
		rows.Scan(&a.ID, &a.GivenName, &a.FamilyName)
		authors = append(authors, a)
	}
	if rows.Err() != nil {
		return []model.Author{}, fmt.Errorf("error when collecting next row: %w", err)
	}

	return authors, nil
}
