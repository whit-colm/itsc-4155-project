package db

import (
	"context"
	"embed"
	"sync"
)

var migrationsFS embed.FS

// Because of how the pathing works, we actually have to get the main
// func to do the embed, the first thing it does in its init method is
// pass the embedded data here; we use once to make sure it doesn't get
// touched after this. It's jank but what're you gonna do
var setOnce sync.Once

func SetMigrationsFS(fs embed.FS) {
	setOnce.Do(func() {
		migrationsFS = fs
	})
}

func (pg *postgres) Migrate(ctx context.Context) error {

	return nil
}
