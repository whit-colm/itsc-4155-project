package cmd

import (
	"flag"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/whit-colm/itsc-4155-project/pkg/endpoints"
)

type flagVars struct {
	GinHost string
	GinPort int

	PsqlDatabase string
	PsqlUser     string
	PsqlPassword string
	PsqlHost     string
	PsqlPort     int
}

var runtimeConfig flagVars

func Run(args []string) int {
	// Define a bunch of flags and parse them
	flag.StringVar(&runtimeConfig.GinHost, "-host", "localhost", "Hostname to listen to")
	flag.IntVar(&runtimeConfig.GinPort, "-port", 8080, "Port to listen to")

	flag.StringVar(&runtimeConfig.PsqlPassword, "-dbdatabase", "postgres", "Database to be used in the PostgreSQL instance")
	flag.StringVar(&runtimeConfig.PsqlUser, "-dbuser", "postgres", "Username for the PostgreSQL user")
	flag.StringVar(&runtimeConfig.PsqlDatabase, "-dbpasswd", "", "Password for the PostgreSQL user")
	flag.StringVar(&runtimeConfig.PsqlHost, "-dbhost", "", "Hostname or IP for the PostgeSQL instance")
	flag.IntVar(&runtimeConfig.PsqlPort, "-dbport", 5432, "Port for the PostgreSQL instance")

	flag.Parse()

	// Instantiate a database connection
	// TODO: this [make sure it's a goroutine]

	// Define the Gin router and endpoints
	router := gin.Default()
	// TODO: externalize *this* as well.
	router.GET("/albums", endpoints.GetAlbums)

	// Start the router
	router.Run(fmt.Sprintf("%v:%d", runtimeConfig.GinHost, runtimeConfig.GinPort))
	return 0
}
