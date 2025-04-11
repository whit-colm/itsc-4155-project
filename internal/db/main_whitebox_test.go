package db

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

var (
	ar authorRepository[string]
	br bookRepository[string]
	pg *postgres
)

// We do this to trick defers into working. mostly.
func wrapTestMain(ctx context.Context, repos []repository.DummyPopulator, m *testing.M) error {
	// Not all devices are equipped to test the DB code, and that's ok
	// (well not really, but we're too poor and strapped for time to do
	// anything). So we check a db uri from ENV vars, if one exists we
	// use it and fail tests; otherwise we "pass" and skip the whole
	// sordid affair.
	uriString := os.Getenv("DB_URI")
	if uriString == "" {
		fmt.Printf("skipping tests; empty `DB_URI` variable.\n")
		return nil
	}
	pg = &postgres{}

	err := pg.Connect(ctx, uriString)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	defer pg.Disconnect()

	for _, r := range repos {
		if err := r.SetDatastore(ctx, pg.db); err != nil {
			return fmt.Errorf("could not set datastore: %w", err)
		}
	}

	p := false
	for _, v := range repos {
		if !v.IsPrepopulated(ctx) {
			if err := v.PopulateDummyValues(ctx); err != nil {
				return fmt.Errorf("failed to instantiate dummy values for %v: %w",
					v, err)
			}
			p = true
		} else {
			v.CleanDummyValues(ctx)
			return fmt.Errorf("database table %v is already populated. cleaning and stopping",
				v)
		}
	}

	time.Sleep(10 * time.Second)
	code := m.Run()

	// This was going to be a defer but it doesn't work for some reason.
	if p {
		for _, v := range repos {
			if err := v.CleanDummyValues(context.Background()); err != nil {
				return fmt.Errorf("failed to clean dummy values: %w", err)
			}
		}
	}

	if code != 0 {
		return fmt.Errorf("nonzero exit code from tests: %d", code)
	}
	return nil
}

// I was going to do this with flags, but tests and flags are real
// gnarly together, so we instead pass an env var. Yes I know this has
// Issues:tm: with Windows
func TestMain(m *testing.M) {
	ctx := context.Background()
	var repos = []repository.DummyPopulator{
		&ar, &br,
	}

	if err := wrapTestMain(ctx, repos, m); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

// TODO: refactor into own file
func TestPing(t *testing.T) {
	if err := pg.db.Ping(t.Context()); err != nil {
		t.Errorf("database failed ping: %s", err)
	}
}
