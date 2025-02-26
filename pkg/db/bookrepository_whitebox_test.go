package db

import (
	"context"

	//"flag"
	"fmt"
	"os"
	"testing"
)

var br bookRepository

// I was going to do this with flags, but tests and flags are real
// ganrly together, so we instead pass an env var. Yes I know this has
// Issues:tm: with Windows
func TestMain(m *testing.M) {
	// Not all devices are
	uriString := os.Getenv("DB_URL")
	if uriString == "" {
		fmt.Printf("skipping tests; empty DB_URL variable.\n")
		os.Exit(0)
	}

	c, err := New(uriString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to PostgreSQL: %v\n", err)
		os.Exit(1)
	}
	defer c.Disconnect()
	br.db = c.db

	os.Exit(m.Run())
}

func TestPing(t *testing.T) {
	if err := br.db.Ping(context.Background()); err != nil {
		t.Errorf("database failed ping: %s", err)
	}
}
