package db

import (
	"context"
	"encoding/json"
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
	tx, err := b.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO books (id, title, published)
		 VALUES ($1, $2, $3)`,
		book.ID, book.Title, book.Published.In(time.UTC),
	)
	if err != nil {
		return fmt.Errorf("create book: %w", err)
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

func (b *bookRepository[S]) getWhere(ctx context.Context, clause string, vals ...any) (*model.Book, error) {
	var book model.Book
	var published time.Time
	var isbns []byte
	var authorIDs []byte

	query := fmt.Sprintf(`SELECT 
			 b.id, 
			 b.title, 
			 b.published,
			 COALESCE(
				 json_agg(a.id) FILTER (WHERE a.id IS NOT NULL),
				 '[]'::json
			 ),
			 COALESCE(
				 json_agg(json_build_object(
					 'value', i.isbn,
					 'type', i.isbn_type
				 )) FILTER (WHERE i.isbn IS NOT NULL),
				 '[]'::json
			 )
		 FROM books b
		 LEFT JOIN books_authors ba ON b.id = ba.book_id
		 LEFT JOIN isbns i ON b.id = i.book_id
		 WHERE %v,
		 GROUP BY b.id`,
		clause)
	if err := b.db.QueryRow(ctx,
		query, vals...,
	).Scan(
		&book.ID, &book.Title, &published, &authorIDs, &isbns,
	); err != nil {
		return nil, fmt.Errorf("get book: %w", err)
	}

	book.Published = civil.DateOf(published)

	if err := json.Unmarshal(isbns, &book.ISBNs); err != nil {
		return nil, fmt.Errorf("get book: %w", err)
	}
	if err := json.Unmarshal(authorIDs, &book.AuthorIDs); err != nil {
		return nil, fmt.Errorf("get book: %w", err)
	}

	return &book, nil
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
func (b *bookRepository[S]) Search(ctx context.Context, offset int, limit int, query ...string) ([]repository.SearchResult[model.BookSummary], error) {
	const errorCaller string = "book search"
	var results []repository.SearchResult[model.BookSummary]

	qStr := strings.Join(query, " ")
	rows, err := b.db.Query(ctx,
		`SELECT
			 paradedb.score(b.id),
		     b.id,
			 b.title,
			 v.published,
			 v.authors,
			 v.isbns,
		 FROM books b
		 LEFT JOIN v_books_summary ON c.poster_id = u.id
	 	 WHERE b.title @@@ $1
		 ORDER BY paradedb.score(b.id) DESC, v.title DESC
		 LIMIT $2 OFFSET $3`,
		qStr,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", errorCaller, err)
	}

	for rows.Next() {
		var (
			s  float64
			o  model.BookSummary
			aS []byte
			iS []byte
		)

		if err = rows.Scan(
			&s, &o.ID, &o.Title, &o.Published, &aS, &iS,
		); err != nil {
			return nil, fmt.Errorf("%v: %w", errorCaller, err)
		}

		if err = json.Unmarshal(aS, &o.Authors); err != nil {
			return nil, fmt.Errorf("%v: %w", errorCaller, err)
		} else if err = json.Unmarshal(iS, &o.ISBNs); err != nil {
			return nil, fmt.Errorf("%v: %w", errorCaller, err)
		}

		r := repository.SearchResult[model.BookSummary]{
			Item:  &o,
			Score: s,
		}
		results = append(results, r)
	}

	return results, rows.Err()
}

// Summarize implements repository.BookManager.
func (b *bookRepository[S]) Summarize(context.Context, *model.Book) (*model.BookSummary, error) {
	panic("unimplemented")
}
