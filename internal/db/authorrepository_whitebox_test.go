package db

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whit-colm/itsc-4155-project/internal/testhelper/dummyvalues"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

/** Implement DummyPopulator **/

// Useful to check that a type implements an interface
var _ repository.DummyPopulator = (*authorRepository)(nil)

func (a *authorRepository) PopulateDummyValues(ctx context.Context) error {
	batch := &pgx.Batch{}

	for _, author := range dummyvalues.ExampleAuthors {
		batch.Queue(`INSERT INTO authors (id, givenname, familyname)
					 VALUES ($1, $2, $3)`,
			author.ID, author.GivenName, author.FamilyName)
	}

	results := a.db.SendBatch(ctx, batch)
	defer results.Close()

	return nil
}

func (a *authorRepository) IsPrepopulated(ctx context.Context) bool {
	var ids uuid.UUIDs
	for _, v := range dummyvalues.ExampleAuthors {
		ids = append(ids, v.ID)
	}
	var count int
	if err := a.db.QueryRow(ctx,
		`SELECT
			 COUNT(*)
		 FROM authors a
		 WHERE a.id = any($1)`,
		ids).Scan(&count); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return false
	}
	return count == len(ids)
}

func (a *authorRepository) CleanDummyValues(ctx context.Context) error {
	tx, err := a.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err = a.db.Exec(ctx, `TRUNCATE authors CASCADE`); err != nil {
		return fmt.Errorf("could not truncate: %w", err)
	}
	return tx.Commit(ctx)
}

func (a *authorRepository) SetDatastore(ctx context.Context, ds any) error {
	if db, ok := ds.(*pgxpool.Pool); !ok {
		return fmt.Errorf("unable to cast ds into pgxpool Pool.")
	} else {
		a.db = db
		return nil
	}
}

/** Actual tests below **/

func TestAuthorCreate(t *testing.T) {

}
