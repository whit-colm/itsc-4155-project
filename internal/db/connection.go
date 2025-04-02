package db

import (
	"context"
	"crypto"
	"crypto/ed25519"
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
func (p *postgres) Connect(ctx context.Context, args ...any) error {
	uri, chn, err := func(args ...any) (string, chan<- error, error) {
		if len(args) != 2 {
			panic("WRONG LEN")
			return "", nil, fmt.Errorf("invalid number of arguments, want `2` have `%d`",
				len(args),
			)
		}
		uriA := args[0]
		chnA := args[1]
		uri, ok := uriA.(string)
		if !ok {
			panic("BAD CAST URI")
			return "", nil, fmt.Errorf("cannot cast arg uri (`%#v`) to `string`", uriA)
		}
		chn, ok := chnA.(chan error)
		if !ok {
			panic("BAD CAST CHAN")
			return "", nil, fmt.Errorf("cannot cast arg eCh (`%#v`) to `chan error`", chnA)
		}
		return uri, chn, nil
	}(args...)
	if err != nil {
		return fmt.Errorf("parse args: %w", err)
	}
	defer close(chn)

	// Connect & Ping the server or die trying.
	for {
		p.db, err = pgxpool.New(ctx, uri)
		if err != nil {
			errF := fmt.Errorf("connect to db: %w", err)
			chn <- errF
			return errF
		}
		// We do not care (ish) about the error, we just keep trying
		// until it wors or expires
		// TODO: care about the error
		if err = p.Ping(ctx); err != nil {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}

	//* first boot & startup maintnenace from here on

	// The rest of this is initial db setup, inserting
	// necessary fields into the database like cryptographic keys
	if _, _, exp, err := p.keyData(ctx); err != nil || time.Now().After(exp) {
		const YEAR = (365 * 24 * time.Hour)
		if _, err = p.Rotate(ctx, YEAR); err != nil {
			errF := fmt.Errorf("rotate keys from key err: %w", err)
			chn <- errF
			return errF
		}
	}

	// Check blob cache table configuration
	var jsonData []byte
	var cacheConfigured bool
	if err = p.db.QueryRow(ctx,
		`SELECT EXISTS (
			 SELECT 1 FROM admin WHERE key = 'blobs_cache_config'
		 ),
		 (
		 	 SELECT value FROM admin WHERE key = 'blobs_cache_config'
		 )`,
	).Scan(&cacheConfigured, &jsonData); err != nil && !errors.Is(err, pgx.ErrNoRows) {

		errF := fmt.Errorf("retrieve key config: %w", err)
		chn <- errF
		return errF
	}

	needsRegen := !cacheConfigured

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
			errF := fmt.Errorf("update blobs_cache_config: %w", err)
			chn <- errF
			return errF
		}
	}
	return nil
}

func NewRepository(uri string, timeout time.Duration) (repository.Repository, error) {
	db := &postgres{}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	eCh := make(chan error)

	go connOnce.Do(func() {
		db.Connect(ctx, uri, eCh)
	})

	var r repository.Repository
	select {
	case err := <-eCh:
		if err != nil {
			return r, fmt.Errorf("instantiate repository: %w", err)
		}
	case <-ctx.Done():
		return r, fmt.Errorf("%w", ctx.Err())
	}

	r.Store = db
	r.Auth = db
	r.Book = newBookRepository(db)
	r.Author = newAuthorRepository(db)
	r.User = newUserRepository(db)
	r.Blob = newBlobRepository(db)
	return r, nil
}

// Disconnect the connection.
//
// If one has not been established, this will do nothing.
func (pg *postgres) Disconnect() error {
	pg.db.Close()
	return nil
}

func (pg *postgres) keyData(ctx context.Context) (ed25519.PrivateKey, ed25519.PublicKey, time.Time, error) {
	var authExists bool
	var jsonData []byte
	if err := pg.db.QueryRow(ctx,
		`SELECT EXISTS (
			 SELECT 1 FROM admin WHERE key = 'auth_crypto'
		 ),
		 (
		 	 SELECT value FROM admin WHERE key = 'auth_crypto'
		 )`,
	).Scan(&authExists, &jsonData); err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, time.Time{}, fmt.Errorf("retrieve key config: %w", err)
	}

	if !authExists {
		return nil, nil, time.Time{}, fmt.Errorf("key not found")
	}

	// Keys stored in b64
	var aux struct {
		Private string    `json:"priv"`
		Public  string    `json:"pub"`
		Expiry  time.Time `json:"expiry"`
	}
	if err := json.Unmarshal(jsonData, &aux); err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("unmarshal key: %w", err)
	}
	var privKey ed25519.PrivateKey
	var pubKey ed25519.PublicKey
	var err error
	privKey, err = base64.StdEncoding.DecodeString(aux.Private)
	if err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("decode key: %w", err)
	}
	pubKey, err = base64.StdEncoding.DecodeString(aux.Public)
	if err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("decode key: %w", err)
	}
	return privKey, pubKey, aux.Expiry, nil
}

func (pg *postgres) KeyPair(ctx context.Context) (crypto.PublicKey, crypto.Signer, error) {
	priv, pub, _, err := pg.keyData(ctx)
	return pub, priv, err
}

func (pg *postgres) Public(ctx context.Context) (crypto.PublicKey, error) {
	_, pub, _, err := pg.keyData(ctx)
	return pub, err
}

func (pg *postgres) Expiry(ctx context.Context) (time.Time, error) {
	_, _, exp, err := pg.keyData(ctx)
	return exp, err
}

func (pg *postgres) Rotate(ctx context.Context, ttl time.Duration) (crypto.Signer, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	privS := base64.StdEncoding.EncodeToString(priv)
	pubS := base64.StdEncoding.EncodeToString(pub)
	if err != nil {
		return nil, err
	}
	e := time.Now().Add(ttl)

	if _, err := pg.db.Exec(ctx,
		`INSERT INTO admin (key, value)
	 	 VALUES ('auth_crypto', jsonb_build_object(
	 	 	 'priv', $1::TEXT,
			 'pub', $2::TEXT,
		 	 'expiry', $3::TIMESTAMPTZ
	 	 ))
	 	 ON CONFLICT (key) DO 
		 UPDATE SET value = jsonb_build_object(
		 	 'priv', $1::TEXT,
			 'pub', $2::TEXT,
		 	 'expiry', $3::TIMESTAMPTZ
	 	 )`,
		privS,
		pubS,
		e.Format(time.RFC3339),
	); err != nil {
		return nil, fmt.Errorf("update auth_crypto: %w", err)
	}
	return priv, nil
}
