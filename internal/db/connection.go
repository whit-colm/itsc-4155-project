package db

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
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
func (p *postgres) Connect(args any) error {
	uri, ok := args.(string)
	if !ok {
		return fmt.Errorf("failure casting args to URI: %#v", args)
	}

	var err error
	connOnce.Do(func() {
		p.db, err = pgxpool.New(context.Background(), uri)
	})
	if err != nil {
		fmt.Errorf("instantiate db pool: %w", err)
	}

	// The rest of this is initial db setup, inserting
	// necessary fields into the database like cryptographic keys
	var ctx context.Context = context.Background()
	var authExists bool
	var jsonData []byte
	if err = p.db.QueryRow(ctx,
		`SELECT EXISTS (
			 SELECT 1 FROM admin WHERE key = 'auth_crypto'
		 ),
		 (
		 	 SELECT value FROM admin WHERE key = 'auth_crypto'
		 )`,
	).Scan(&authExists, &jsonData); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		fmt.Errorf("retrieve key config: %w", err)
	}

	needsRegen := !authExists
	genSecret := func() string {
		secret := make([]byte, 96)
		rand.Read(secret)
		return base64.StdEncoding.EncodeToString(secret)
	}

	if authExists {
		var aux struct {
			Secret string    `json:"secret"`
			Expiry time.Time `json:"expiry"`
		}
		if err := json.Unmarshal(jsonData, &aux); err != nil {
			// If json is invalid, we just regenerate it
			needsRegen = true
		} else {
			// Check conditions that require regeneration:
			// 1. Key is empty
			// 2. Expiry is empty
			// 3. The current time is after expiry
			needsRegen = aux.Secret == "" ||
				aux.Expiry.IsZero() ||
				time.Now().After(aux.Expiry)
		}
	}

	if needsRegen {
		s := genSecret()
		e := time.Now().Add(365 * 24 * time.Hour)

		if _, err := p.db.Exec(ctx,
			`INSERT INTO admin (key, value)
			 VALUES ('auth_crypto', jsonb_build_object(
			 	 'secret', $1::TEXT,
				 'expiry', $2::TIMESTAMPTZ
			 ))
			 ON CONFLICT (key) DO 
			 UPDATE SET value = jsonb_build_object(
			 	 'secret', $1::TEXT,
				 'expiry', $2::TIMESTAMPTZ
			 )`, s, e.Format(time.RFC3339),
		); err != nil {
			return fmt.Errorf("update auth_crypto: %w", err)
		}
	}

	// Check blob cache table configuration
	var cacheConfigured bool
	if err = p.db.QueryRow(ctx,
		`SELECT EXISTS (
			 SELECT 1 FROM admin WHERE key = 'blobs_cache_config'
		 ),
		 (
		 	 SELECT value FROM admin WHERE key = 'blobs_cache_config'
		 )`,
	).Scan(&cacheConfigured, &jsonData); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("retrieve key config: %w", err)
	}

	needsRegen = !cacheConfigured

	if cacheConfigured {
		var aux struct {
			MCS int           `json:"maxCache"`
			TTL time.Duration `json:"ttl"`
		}
		if err := json.Unmarshal(jsonData, &aux); err != nil {
			// If json is invalid, we just regenerate it
			needsRegen = true
		} else {
			// Check conditions that require regeneration:
			// 1. Max Cache Size is 0
			// 2. TTL is Zero
			needsRegen = aux.MCS == 0 ||
				aux.TTL == 0
		}
	}

	if needsRegen {
		// TODO: no magic numbers.
		// This is 3/4 of a GiB != 750 MiB (yay base2)
		// or (1<<29) + (1<<28)
		const maxCache uint = 805306368
		const ttl time.Duration = (1 * time.Hour)
		if _, err := p.db.Exec(ctx,
			`INSERT INTO admin (key, value)
			 VALUES ('blobs_cache_config', jsonb_build_object(
			 	 'maxSize', $1::BIGINT,
				 'ttl', ($2::INTEGER::TEXT || ' seconds')::INTERVAL
			 ))
			 ON CONFLICT (key) DO 
			 UPDATE SET value = jsonb_build_object(
			 	 'maxSize', $1::BIGINT,
				 'ttl', ($2::INTEGER::TEXT || ' seconds')::INTERVAL
			 )`, maxCache, ttl.Seconds(),
		); err != nil {
			return fmt.Errorf("update blobs_cache_config: %w", err)
		}
	}
	return nil
}

func NewRepository(uri string) (r repository.Repository, err error) {
	db := &postgres{}
	err = db.Connect(uri)
	r.Store = db
	r.Book = newBookRepository(db)
	r.Author = newAuthorRepository(db)
	r.User = newUserRepository(db)
	r.Blob = newBlobRepository(db)
	return
}

// Disconnect the connection.
//
// If one has not been established, this will do nothing.
func (pg *postgres) Disconnect() error {
	pg.db.Close()
	return nil
}
