package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (a authorRepository[S]) queryString(clause string, search bool) string {
	return fmt.Sprintf(`SELECT
			 %v
			 a.id,
			 COALESCE(a.given_name, ''),
			 a.family_name,
			 COALESCE(a.bio, ''),
			 COALESCE(
			 	 json_agg(json_build_object(
				 	 'type', i.type,
					 identifier, i.identifier
			 	 )) FILTER (WHERE i.author_id IS NOT NULL),
				 '[]'::json
			 )
		 FROM authors a
		 LEFT JOIN author_identifiers i ON i.author_id = a.id
		 WHERE %v`,
		func() string {
			if search {
				return "paradedb.score(a.id),"
			}
			return ""
		}(),
		// The order for select is WHERE -> GROUP BY -> ORDER BY
		// We can't easily arbitrarily shimmy our GROUP BY in the
		// middle so we have to hope the search is correct and
		// insert it depending on that.
		func() string {
			const groupByClause string = `GROUP BY a.id`
			if search {
				return clause
			}
			return clause + "\n" + groupByClause
		}(),
	)
}

func (a authorRepository[S]) rowsParse(rows pgx.Rows, search bool) (*model.Author, float64, error) {
	var (
		author model.Author
		e      []byte
		s      float64
	)

	if search {
		if err := rows.Scan(
			&s, &author.ID, &author.GivenName, &author.FamilyName,
			&author.Bio, &e,
		); err != nil {
			return nil, -1.0, err
		}
	} else {
		// If not searching, we don't have a score
		if err := rows.Scan(
			&author.ID, &author.GivenName, &author.FamilyName,
			&author.Bio, &e,
		); err != nil {
			return nil, -1.0, err
		}
	}

	if err := json.Unmarshal(e, &author.ExtIDs); err != nil {
		return nil, -1.0, err
	}

	return &author, s, nil
}

func (a *authorRepository[S]) Book(ctx context.Context, bookID uuid.UUID) ([]*model.Author, error) {
	panic("unimplemented")
}

// Create implements repository.AuthorManager.
func (a *authorRepository[S]) Create(ctx context.Context, author *model.Author) error {
	if author.FamilyName == "" {
		return fmt.Errorf("author family name cannot be empty")
	}

	tx, err := a.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO authors (id, family_name)
		 VALUES ($1, $2)`,
		author.ID, author.FamilyName,
	)
	if err != nil {
		return fmt.Errorf("failed to insert author: %w", err)
	}
	if author.GivenName != "" {
		_, err = tx.Exec(ctx,
			`UPDATE authors
			 SET given_name = $1
			 WHERE id = $2`,
			author.GivenName, author.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to update author given name: %w", err)
		}
	}
	if author.Bio != "" {
		_, err = tx.Exec(ctx,
			`UPDATE authors
			 SET bio = $1
			 WHERE id = $2`,
			author.Bio, author.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to update author bio: %w", err)
		}
	}

	for _, extID := range author.ExtIDs {
		_, err = tx.Exec(ctx,
			`INSERT INTO author_identifiers (author_id, type, identifier)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (author_id, type, identifier) DO NOTHING`,
			author.ID, extID.Type, extID.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to insert author external identifier: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (a *authorRepository[S]) ExistsByName(ctx context.Context, name string) (*model.Author, bool, error) {
	const errorCaller string = "get author by name"
	var author *model.Author
	rows, err := a.db.Query(ctx,
		a.queryString(
			`TRIM((given_name || ' ' || family_name)) = $1`,
			false,
		),
		name,
	)
	if err != nil {
		fmt.Println("I doubt you will see this error, but if you do, please report it.")
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, repository.Err{Code: repository.ErrNotFound, Err: err}
		}
		return nil, false, fmt.Errorf("%v: %w", errorCaller, err)
	}
	defer rows.Close()

	if rows.Next() {
		author, _, err = a.rowsParse(rows, false)
		if err != nil {
			return nil, false, fmt.Errorf("%v: %w", errorCaller, err)
		}
	} else {
		if errors.Is(rows.Err(), pgx.ErrNoRows) {
			return nil, false, repository.Err{Code: repository.ErrNotFound, Err: rows.Err()}
		}
		return nil, false, rows.Err()
	}
	return author, author != nil && author.ID != uuid.Nil, rows.Err()
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
	const errorCaller string = "author by id"

	var author *model.Author
	rows, err := a.db.Query(ctx,
		a.queryString(`a.id = $1`, false),
		id,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.Err{Code: repository.ErrNotFound, Err: err}
		}
		return nil, fmt.Errorf("%v: %w", errorCaller, err)
	}
	defer rows.Close()

	if rows.Next() {
		author, _, err = a.rowsParse(rows, false)
		if err != nil {
			return nil, fmt.Errorf("%v: %w", errorCaller, err)
		}
	} else {
		// If there's no next, that means there's either an error or no rows
		// check if there's an error and if not just return a repository.ErrNotFound
		if rows.Err() != nil {
			return nil, fmt.Errorf("%v: %w", errorCaller, rows.Err())
		}
		return nil, repository.Err{Code: repository.ErrNotFound, Err: rows.Err()}
	}
	return author, rows.Err()
}

// Search implements repository.AuthorManager.
func (a *authorRepository[S]) Search(ctx context.Context, offset int, limit int, query ...string) ([]repository.SearchResult[model.Author], []repository.AnyScoreItemer, error) {
	const errorCaller string = "author search"
	var resultsT []repository.SearchResult[model.Author]
	var resultsASI []repository.AnyScoreItemer

	qStr := strings.Join(query, " ")
	rows, err := a.db.Query(ctx,
		a.queryString(`family_name @@@ $1 OR given_name @@@ $1
			 GROUP BY a.id
			 ORDER BY paradedb.score(id) DESC, family_name DESC
			 LIMIT $2 OFFSET $3`,
			true,
		),
		qStr,
		limit,
		offset,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("%v: %w", errorCaller, err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			s float64
			u *model.Author
		)

		u, s, err = a.rowsParse(rows, true)
		if err != nil {
			return nil, nil, fmt.Errorf("%v: %w", errorCaller, err)
		}

		r := repository.SearchResult[model.Author]{
			Item:  u,
			Score: s,
		}
		resultsT = append(resultsT, r)
		resultsASI = append(resultsASI, r)
	}

	return resultsT, resultsASI, rows.Err()
}
