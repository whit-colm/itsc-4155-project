package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type authorRepository[S comparable] struct {
	db *pgxpool.Pool
}

// Useful to check that a type implements an interface
var _ repository.AuthorManager[string] = (*authorRepository[string])(nil)

func newAuthorRepository(psql *postgres) repository.AuthorManager[string] {
	return &authorRepository[string]{db: psql.db}
}

func (a *authorRepository[S]) Book(ctx context.Context, bookID uuid.UUID) ([]*model.Author, error) {
	panic("unimplemented")
}

// Create implements repository.AuthorManager.
func (a *authorRepository[S]) Create(ctx context.Context, author *model.Author) error {
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO authors (id, familyname, givenname)
		 VALUES ($1, $2, $3)`,
		author.ID, author.GivenName, author.GivenName,
	)
	if err != nil {
		return fmt.Errorf("failed to insert author: %w", err)
	}

	return tx.Commit(ctx)
}

func (a *authorRepository[S]) ExistsByName(ctx context.Context, name string) (*model.Author, bool, error) {
	const errorCaller string = "get author by name"
	var author model.Author
	if err := a.db.QueryRow(ctx,
		`SELECT
			 id,
			 COALESCE(givenname, '') as gn,
			 familyname
		FROM authors
		WHERE TRIM((gn || ' ' || familyname)) = $1`,
		name,
	).Scan(
		&author.ID, &author.GivenName, &author.FamilyName,
	); err != nil {
		return nil, false, fmt.Errorf("%v: %w", errorCaller, err)
	}
	return &author, author.ID != uuid.Nil, nil
}

// Update implements repository.BookManager.
func (a *authorRepository[S]) Update(ctx context.Context, to *model.Author) (*model.Author, error) {
	panic("unimplemented")
}

// Delete implements repository.AuthorManager.
func (a *authorRepository[S]) Delete(ctx context.Context, id uuid.UUID) error {
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
func (a *authorRepository[S]) GetByID(ctx context.Context, id uuid.UUID) (*model.Author, error) {
	var author model.Author

	if err := a.db.QueryRow(ctx,
		`SELECT
			a.id,
			COALESCE(givenname, ''),
			a.familyname
		FROM authors a
		WHERE a.id = $1`,
		id,
	).Scan(
		&author.ID, &author.GivenName, &author.FamilyName,
	); err != nil {
		return nil, fmt.Errorf("could not retrieve author: %w", err)
	}
	return &author, nil
}

// Search implements repository.AuthorManager.
func (a *authorRepository[S]) Search(ctx context.Context, offset int, limit int, query ...string) ([]repository.SearchResult[model.Author], []repository.AnyScoreItemer, error) {
	const errorCaller string = "book search"
	var resultsT []repository.SearchResult[model.Author]
	var resultsASI []repository.AnyScoreItemer

	qStr := strings.Join(query, " ")
	rows, err := a.db.Query(ctx,
		`SELECT
			 paradedb.score(id),
			 id,
			 family_name,
			 COALESCE(givenname, '')
		 FROM authors
	 	 WHERE family_name @@@ $1 OR given_name @@@ $1
		 ORDER BY paradedb.score(id) DESC, family_name DESC
		 LIMIT $2 OFFSET $3`,
		qStr,
		limit,
		offset,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("%v: %w", errorCaller, err)
	}

	for rows.Next() {
		var (
			s float64
			u model.Author
		)

		if err = rows.Scan(
			&s, &u.ID, &u.FamilyName, &u.GivenName,
		); err != nil {
			return nil, nil, fmt.Errorf("%v: %w", errorCaller, err)
		}

		r := repository.SearchResult[model.Author]{
			Item:  &u,
			Score: s,
		}
		resultsT = append(resultsT, r)
		resultsASI = append(resultsASI, r)
	}

	return resultsT, resultsASI, rows.Err()
}
