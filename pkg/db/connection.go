package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgres struct {
	db  *pgxpool.Pool
	err error
}

func (pg *postgres) Ping() error {
	return pg.db.Ping(context.Background())
}

var connection *postgres = &postgres{}
var connOnce sync.Once

// Establish database connection.
//
// Connect takes a string which should be a valid PostgeSQL URI and attempts to
// establish a connection to the database located at that URL. If successful,
// connection details will be stored in a package-wide private var and used by
// all other methods.
//
// Make sure to defer *Disconnect()* after connecting.
func Connect(uri string) (db *pgxpool.Pool, err error) {
	connOnce.Do(func() {
		db, err := pgxpool.New(context.Background(), uri)
		if err != nil {
			connection.err = fmt.Errorf("failed to parse URI: `%w`", err)
			return
		}
		connection.db = db
	})
	err = connection.err
	db = connection.db
	connection.err = nil

	return
}

// Disconnect the connection.
//
// If one has not been established, this will do nothing.
func Disconnect() {
	connection.db.Close()
}
