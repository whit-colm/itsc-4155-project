package main

import (
	"embed"
	"os"

	"github.com/whit-colm/itsc-4155-project/cmd"
	"github.com/whit-colm/itsc-4155-project/pkg/db"
)

// This is how we handle database migrations, which feels kinda gross
// but what can you do.
//
//go:embed build/migrations/*.sql
var databaseMigrationFiles embed.FS

func main() {
	db.SetMigrationsFS(databaseMigrationFiles)
	os.Exit(cmd.Run(os.Args[1:]))
}
