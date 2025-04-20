package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/civil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type bookRepository[S comparable] struct {
	db *pgxpool.Pool
}

// Useful to check that a type implements an interface
var _ repository.BookManager[string] = (*bookRepository[string])(nil)

func newBookRepository(psql *postgres) repository.BookManager[string] {
	return &bookRepository[string]{db: psql.db}
}

func (b *bookRepository[S]) Author(ctx context.Context, authorID uuid.UUID) ([]*model.Book, error) {
	panic("unimplemented")
}

// Create implements BookRepositoryManager.
func (b *bookRepository[S]) Create(ctx context.Context, book *model.Book) error {
	const errorCaller string = "create book"
	tx, err := b.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO books (id, title, subtitle, description, published)
		 VALUES ($1, $2, $3, $4, $5)`,
		book.ID, book.Title, book.Subtitle, book.Description,
		book.Published.In(time.UTC),
	)
	if err != nil {
		return fmt.Errorf("%v: %w", errorCaller, err)
	}

	// Insert cover and thumbnail images if they are set
	if book.CoverImage != uuid.Nil {
		_, err = tx.Exec(ctx,
			`UPDATE books SET cover_image = $1 WHERE id = $2`,
			book.CoverImage, book.ID,
		)
		if err != nil {
			return fmt.Errorf("%v: %w", errorCaller, err)
		}
	}
	if book.ThumbImage != uuid.Nil {
		_, err = tx.Exec(ctx,
			`UPDATE books SET thumbnail_image = $1 WHERE id = $2`,
			book.ThumbImage, book.ID,
		)
		if err != nil {
			return fmt.Errorf("%v: %w", errorCaller, err)
		}
	}

	rows := [][]interface{}{}
	for _, isbn := range book.ISBNs {
		row := []interface{}{isbn.String(), book.ID, isbn.Version().String()}
		rows = append(rows, row)
	}
	if _, err := tx.CopyFrom(ctx,
		pgx.Identifier{"isbns"},
		[]string{"isbn", "book_id", "isbn_type"},
		pgx.CopyFromRows(rows),
	); err != nil {
		return fmt.Errorf("create book: %w", err)
	}

	rows = [][]interface{}{}
	for _, aID := range book.AuthorIDs {
		row := []interface{}{book.ID, aID}
		rows = append(rows, row)
	}
	if _, err := tx.CopyFrom(ctx,
		pgx.Identifier{"books_authors"},
		[]string{"book_id", "author_id"},
		pgx.CopyFromRows(rows),
	); err != nil {
		return fmt.Errorf("create book: %w", err)
	}

	return tx.Commit(ctx)
}

func (b *bookRepository[S]) ExistsByISBN(ctx context.Context, isbns ...model.ISBN) (*model.Book, bool, error) {
	if len(isbns) == 0 {
		return nil, false, fmt.Errorf("no ISBNs provided")
	}

	// Build query with IN clause
	placeholders := []string{}
	args := []any{}
	for i, isbn := range isbns {
		placeholders = append(placeholders, fmt.Sprintf("$%d", i+1))
		args = append(args, isbn.String())
	}
	clause := fmt.Sprintf("i.isbn IN (%s)", strings.Join(placeholders, ","))
	book, err := b.getWhere(ctx, clause, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, repository.Err{Code: repository.ErrNotFound, Err: err}
		}
		return nil, false, fmt.Errorf("get book: %w", err)
	}
	return book, true, nil
}

func (b bookRepository[S]) queryString(clause string, search bool) string {
	return fmt.Sprintf(`SELECT
		 %v
		 b.id,
		 b.title,
		 COALESCE(b.subtitle, ''),
		 COALESCE(b.description, ''),
		 b.published,
		 COALESCE(
			 jsonb_agg(DISTINCT a.author_id) FILTER (WHERE a.author_id IS NOT NULL),
			 '[]'::json
		 ),
		 COALESCE(
			 jsonb_agg(jsonb_build_object(
				 'value', i.isbn,
				 'type', i.isbn_type
			 )) FILTER (WHERE i.isbn IS NOT NULL),
			 '[]'::json
		 ),
		 b.cover_image,
		 b.thumbnail_image
		 FROM books b
		 LEFT JOIN isbns i ON i.book_id = b.id
		 LEFT JOIN books_authors a ON a.book_id = b.id
		 WHERE %v
		 GROUP BY b.id`,
		func() string {
			if search {
				return "paradedb.score(b.id),"
			}
			return ""
		}(),
		clause,
	)
}

func (b bookRepository[S]) rowsParse(rows pgx.Rows, search bool) (*model.Book, float64, error) {
	var (
		book      model.Book
		published time.Time
		authorIDs []byte
		isbns     []byte
		score     float64
	)

	if search {
		if err := rows.Scan(
			&score, &book.ID, &book.Title, &book.Subtitle, &book.Description,
			&published, &authorIDs, &isbns, &book.CoverImage, &book.ThumbImage,
		); err != nil {
			return nil, -1.0, err
		}
	} else {
		if err := rows.Scan(
			&book.ID, &book.Title, &book.Subtitle, &book.Description,
			&published, &authorIDs, &isbns, &book.CoverImage, &book.ThumbImage,
		); err != nil {
			return nil, -1.0, err
		}
	}

	book.Published = civil.DateOf(published)

	if err := json.Unmarshal(isbns, &book.ISBNs); err != nil {
		return nil, -1.0, err
	}
	if err := json.Unmarshal(authorIDs, &book.AuthorIDs); err != nil {
		return nil, -1.0, err
	}

	return &book, score, nil
}

func (b *bookRepository[S]) getWhere(ctx context.Context, clause string, vals ...any) (*model.Book, error) {
	rows, err := b.db.Query(ctx,
		b.queryString(clause, false),
		vals...,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		book, _, err := b.rowsParse(rows, false)
		if err != nil {
			return nil, err
		}
		return book, nil
	}
	return nil, repository.ErrNotFound
}

// Update implements repository.BookManager.
func (b *bookRepository[S]) Update(ctx context.Context, book *model.Book) (*model.Book, error) {
	panic("unimplemented")
}

func (b *bookRepository[S]) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := b.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`DELETE FROM books b
		 WHERE b.id = $1`,
		id,
	); err != nil {
		return fmt.Errorf("failed to delete book: %w", err)
	}

	// We shouldn't have to delete the ISBNs ourselves, on delete they
	// cascade

	return tx.Commit(ctx)
}

// GetByID implements BookRepositoryManager.
func (b *bookRepository[S]) GetByID(ctx context.Context, id uuid.UUID) (*model.Book, error) {
	return b.getWhere(ctx, "b.id = $1", id.String())
}

// GetByISBN implements BookRepositoryManager.
func (b *bookRepository[S]) GetByISBN(ctx context.Context, isbn model.ISBN) (*model.Book, error) {
	return b.getWhere(ctx, "i.isbn = $1", isbn.String())
}

// Search implements BookRepositoryManager.
func (b *bookRepository[S]) Search(ctx context.Context, offset int, limit int, query ...string) ([]repository.SearchResult[model.BookSummary], []repository.AnyScoreItemer, error) {
	const errorCaller string = "book search"
	var resultsT []repository.SearchResult[model.BookSummary]
	var resultsASI []repository.AnyScoreItemer

	qStr := strings.Join(query, " ")
	rows, err := b.db.Query(ctx,
		`SELECT
			 paradedb.score(b.id),
			 b.id,
			 b.title,
			 b.subtitle,
			 b.description,
			 b.published,
			 v.thumbnail_image,
			 v.authors,
			 v.isbns
		 FROM books b
		 LEFT JOIN v_books_summary v ON v.id = b.id
		 WHERE b.title @@@ $1 OR b.subtitle @@@ $1 OR b.description @@@ $1
		 ORDER BY paradedb.score(b.id) DESC, v.title DESC
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
			s  float64
			o  model.BookSummary
			aS []byte
			iS []byte
		)

		if err = rows.Scan(
			&s, &o.ID, &o.Title, &o.Subtitle, &o.Description,
			&o.Published, &o.ThumbImage, &aS, &iS,
		); err != nil {
			return nil, nil, fmt.Errorf("%v: %w", errorCaller, err)
		}

		if err = json.Unmarshal(aS, &o.Authors); err != nil {
			return nil, nil, fmt.Errorf("%v: %w", errorCaller, err)
		} else if err = json.Unmarshal(iS, &o.ISBNs); err != nil {
			return nil, nil, fmt.Errorf("%v: %w", errorCaller, err)
		}

		r := repository.SearchResult[model.BookSummary]{
			Item:  &o,
			Score: s,
		}
		resultsT = append(resultsT, r)
		resultsASI = append(resultsASI, r)
	}

	return resultsT, resultsASI, rows.Err()
}

// Summarize implements repository.BookManager.
func (b *bookRepository[S]) Summarize(context.Context, *model.Book) (*model.BookSummary, error) {
	panic("unimplemented")
}
