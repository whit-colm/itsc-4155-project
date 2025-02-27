package db

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/civil"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type bookRepository struct {
	db *pgxpool.Pool
}

func newBookRepository(psql *postgres) repository.BookManager {
	return &bookRepository{db: psql.db}
}

// Create implements BookRepositoryManager.
func (b *bookRepository) Create(ctx context.Context, book *model.Book) error {
	tx, err := b.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO books (id, title, author, published)
		 VALUES ($1, $2, $3, $4)`,
		book.ID, book.Title, book.Author, book.Published.In(time.UTC),
	)
	if err != nil {
		return fmt.Errorf("failed to insert book: %w", err)
	}

	for _, isbn := range book.ISBNs {
		_, err := tx.Exec(ctx,
			`insert INTO isbns (isbn, book_id, isbn_type)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (isbn) DO NOTHING`,
			isbn.String(), book.ID, isbn.Version().String(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert ISBN %s: %w", isbn, err)
		}
	}

	return tx.Commit(ctx)
}

func (b *bookRepository) Delete(ctx context.Context, book *model.Book) error {
	tx, err := b.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var id uuid.UUID
	if err := tx.QueryRow(ctx,
		`DELETE FROM books b
		 WHERE b.id = $1
		 RETURNING id`,
		book.ID).Scan(&id); err != nil {
		return fmt.Errorf("failed to delete book: %w", err)
	}

	if id != book.ID {
		return fmt.Errorf("wrong book was deleted. want: `%v`; have: `%v`", book.ID, id)
	}

	// We shouldn't have to delete the ISBNs ourselves, on delete they
	// cascade

	return tx.Commit(ctx)
}

// GetByID implements BookRepositoryManager.
func (b *bookRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Book, error) {
	var book model.Book
	var published time.Time
	var isbns []byte

	if err := b.db.QueryRow(ctx,
		`SELECT 
			b.id, 
			b.title, 
			b.author, 
			b.published,
			COALESCE(
				json_agg(json_build_object(
					'value', i.isbn,
					'type', i.isbn_type
				)) FILTER (WHERE i.isbn IS NOT NULL),
				'[]'::json
			) AS isbns
		FROM books b
		LEFT JOIN isbns i ON b.id = i.book_id
		WHERE b.id = $1
		GROUP BY b.id`,
		id,
	).Scan(
		&book.ID, &book.Title, &book.Author,
		&published, &isbns,
	); err != nil {
		return nil, err
	}

	book.Published = civil.DateOf(published)

	if err := json.Unmarshal(isbns, &book.ISBNs); err != nil {
		return nil, fmt.Errorf("failed to parse ISBNs: %w", err)
	}

	return &book, nil
}

// GetByISBN implements BookRepositoryManager.
func (b *bookRepository) GetByISBN(ctx context.Context, isbn model.ISBN) (uuid.UUID, *model.Book, error) {
	var book model.Book
	var published time.Time
	var isbns []byte

	if err := b.db.QueryRow(ctx,
		`SELECT 
			b.id, 
			b.title, 
			b.author, 
			b.published,
			COALESCE(
				json_agg(json_build_object(
					'value', i.isbn,
					'type', i.isbn_type
				)) FILTER (WHERE i.isbn IS NOT NULL),
				'[]'::json
			) AS isbns
		FROM books b
		LEFT JOIN isbns i ON b.id = i.book_id
		WHERE EXISTS (
			SELECT * FROM isbns 
			WHERE isbn = $1 AND book_id = b.id
		)
		GROUP BY b.id`,
		isbn,
	).Scan(
		&book.ID, &book.Title, &book.Author,
		&published, &isbns,
	); err != nil {
		return uuid.Nil, nil, err
	}

	book.Published = civil.DateOf(published)

	if err := json.Unmarshal(isbns, &book.ISBNs); err != nil {
		return uuid.Nil, nil, fmt.Errorf("failed to parse ISBNs: %w", err)
	}

	return book.ID, &book, nil
}

// Search implements BookRepositoryManager.
func (b *bookRepository) Search(ctx context.Context) ([]model.Book, error) {
	rows, err := b.db.Query(ctx,
		`SELECT 
			b.id, 
			b.title, 
			b.author, 
			b.published,
			COALESCE(
				json_agg(json_build_object(
					'value', i.isbn,
					'type', i.isbn_type
				)) FILTER (WHERE i.isbn IS NOT NULL),
				'[]'::json
			) AS isbns
		FROM books b
		LEFT JOIN isbns i ON b.id = i.book_id
		WHERE *
		GROUP BY b.id`)

	if err != nil {
		return []model.Book{}, err
	}
	return pgx.CollectRows(rows, pgx.RowToStructByName[model.Book])
}
