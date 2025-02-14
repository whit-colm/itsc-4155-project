package db

import (
	"context"

	"github.com/jackc/pgx/v5"
)

//TODO: This should NOT be set in stone!!

func Connect(url string) (*pgx.Conn, error) {
	return pgx.Connect(context.Background(), url)
}
