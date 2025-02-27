package db

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type dbError string

func (e dbError) Error() string {
	return string(e)
}

const (
	ConnectionNotEstablished dbError = "Database connection has not been established"
	ConnectionNotCreated     dbError = "Unable to create connection pool"
)

func handlePGError(err error, op string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("`%s` not found: %w", op, err)
	}
	return fmt.Errorf("database error during `%s`: %w", op, err)
}
