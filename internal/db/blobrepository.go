package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type blobRepository struct {
	db *pgxpool.Pool
}

func newBlobRepository(psql *postgres) repository.BlobManager {
	return &blobRepository{db: psql.db}
}

// Useful to check that a type implements an interface
var _ repository.BlobManager = (*blobRepository)(nil)

// Create implements repository.BlobManager.
func (b *blobRepository) Create(ctx context.Context, t *model.Blob) error {
	tx, err := b.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	data, err := io.ReadAll(t.Content)
	if err != nil {
		return fmt.Errorf("read blob content: %w", err)
	}

	if _, err = tx.Exec(ctx,
		`INSERT INTO blobs (id, metadata, value)
		 VALUES ($1, $2, $3)`,
		t.ID, t.Metadata, data,
	); err != nil {
		return fmt.Errorf("insert into blobs: %w", err)
	}

	return tx.Commit(ctx)
}

// Delete implements repository.BlobManager.
func (b *blobRepository) Delete(ctx context.Context, id uuid.UUID) error {
	tx, err := b.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`DELETE FROM blobs
		 WHERE id = $1`,
		id,
	); err != nil {
		return fmt.Errorf("delete blob: %w", err)
	}

	return tx.Commit(ctx)
}

// GetByID implements repository.BlobManager.
//
// This is mostly just a wrapper for the custom cache system
// implemented in the database itself.
func (b *blobRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Blob, error) {
	var blob model.Blob
	var metadata []byte
	var data []byte

	if err := b.db.QueryRow(ctx,
		`SELECT id, metadata, value FROM get_blob($1)`,
		id,
	).Scan(&blob.ID, &metadata, &data); err != nil {
		return nil, fmt.Errorf("receive blob: %w", err)
	}
	blob.Content = bytes.NewReader(data)
	if err := json.Unmarshal(metadata, &blob.Metadata); err != nil {
		return nil, fmt.Errorf("receive blob: %w", err)
	}

	return &blob, nil
}

// Necessary to implement repository.BlobManager.
// This should never be called. If it does it just creates a new blob,
// blobs are considered immutable.
func (b *blobRepository) Update(ctx context.Context, t *model.Blob) (*model.Blob, error) {
	if id, err := uuid.NewV7(); err != nil {
		return nil, fmt.Errorf("hey you shouldn't be using this method (%w)", err)
	} else {
		newBlob := model.Blob{
			ID:      id,
			Content: t.Content,
		}
		return &newBlob, b.Create(ctx, &newBlob)
	}
}

/** Here be dragons
// This is some hacky something I was trying to fit together but I
// think I'm bringing a Chengdu J-20 Mighty Dragon All-Weather Stealth
// fighter to a knife fight. This code might be useful for reuse if
// you're using Postgres LO (large objects) but for Bytea, which is
// significantly smaller it's insane overkill.

type streamingCopySource struct {
	reader io.Reader
	buf    []byte
	chunk  []byte
	done   bool
	err    error
}

func (s *streamingCopySource) Next() bool {
	if s.done {
		return false
	}

	n, err := s.reader.Read(s.buf)
	if errors.Is(err, io.EOF) {
		s.done = true
	} else if err != nil {
		s.err = err
	}
	s.chunk = s.buf[:n]
	return n > 0 || !s.done
}

func (s *streamingCopySource) Values() ([]any, error) {
	return []any{s.chunk}, nil
}

func (s *streamingCopySource) Err() error {
	return s.err
}

// Create implements repository.BlobManager.
func (b *blobRepository) Create(ctx context.Context, t *model.Blob) error {
	conn, err := b.db.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err = conn.Exec(ctx,
		`CREATE TEMPORARY TABLE blobs_tmp (
			 rid SERIAL PRIMARY KEY,
			 id UUID NOT NULL,
			 value BYTEA
		 ) ON COMMIT DROP;`,
	); err != nil {
		return fmt.Errorf("create temp table: %w", err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = conn.CopyFrom(
		ctx,
		pgx.Identifier{"blobs_tmp"},
	)

	return tx.Commit(ctx)
}

// Delete implements repository.BlobManager.
func (b *blobRepository) Delete(ctx context.Context, t *model.Blob) error {
	panic("unimplemented")
}

// GetByID implements repository.BlobManager.
//
// This is mostly just a wrapper for the custom cache system
// implemented in the database itself.
func (b *blobRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Blob, error) {
	panic("unimplemented")
}

// Necessary to implement repository.BlobManager.
// This should never be called.
func (b *blobRepository) Update(ctx context.Context, t *model.Blob) (*model.Blob, error) {
	panic("unimplemented")
}

*/
