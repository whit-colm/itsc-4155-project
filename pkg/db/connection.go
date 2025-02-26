package db

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type postgres struct {
	db *pgxpool.Pool
}

func (pg *postgres) Ping(ctx context.Context) error {
	return pg.db.Ping(ctx)
}

var connOnce sync.Once

// Establish database connection.
//
// Connect takes a string which should be a valid PostgeSQL URI and attempts to
// establish a connection to the database located at that URL. If successful,
// connection details will be stored in a package-wide private var and used by
// all other methods.
//
// Make sure to defer *Disconnect()* after connecting.
func New(uri string) (*postgres, error) {
	s := &postgres{}
	var err error
	connOnce.Do(func() {
		s.db, err = pgxpool.New(context.Background(), uri)
	})

	return s, err
}

func NewRepository(uri string) (r repository.Repository, err error) {
	db, err := New(uri)
	r.Store = db
	r.Book = newBookRepository(db)
	return
}

// Disconnect the connection.
//
// If one has not been established, this will do nothing.
func (pg *postgres) Disconnect() error {
	pg.db.Close()
	return nil
}
