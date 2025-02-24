package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/whit-colm/itsc-4155-project/pkg/db"
	"github.com/whit-colm/itsc-4155-project/pkg/endpoints"
)

type flagVars struct {
	DockerMode bool
	DebugMode  bool

	GinHost string
	GinPort string

	PsqlDatabase string
	PsqlUser     string
	PsqlPassword string
	PsqlHost     string
	PsqlPort     string
}

var runtimeConfig flagVars

func Run(args []string) int {
	// Define a bunch of flags and parse them
	flag.BoolVar(&runtimeConfig.DockerMode, "docker", false, "Weather to run in Docker mode (i.e. read ENV vars)")
	flag.BoolVar(&runtimeConfig.DebugMode, "debug", true, "Weather to run in Docker mode (i.e. read ENV vars)")
	flag.StringVar(&runtimeConfig.GinHost, "host", "localhost", "Hostname to listen to")
	flag.StringVar(&runtimeConfig.GinPort, "port", "9000", "Port to listen to")

	flag.StringVar(&runtimeConfig.PsqlPassword, "dbdatabase", "jaws", "Database to be used in the PostgreSQL instance")
	flag.StringVar(&runtimeConfig.PsqlUser, "dbuser", "jaws", "Username for the PostgreSQL user")
	flag.StringVar(&runtimeConfig.PsqlDatabase, "dbpasswd", "", "Password for the PostgreSQL user")
	flag.StringVar(&runtimeConfig.PsqlHost, "dbhost", "127.0.0.1", "Hostname or IP for the PostgeSQL instance")
	flag.StringVar(&runtimeConfig.PsqlPort, "dbport", "5432", "Port for the PostgreSQL instance")

	flag.Parse()

	// Before continuing, check if running in docker mode
	// if we are we need to re-read runtime values from env vars
	if runtimeConfig.DockerMode {
		// In docker mode, the backend is only ever touched by the
		// front-end (nginx); and is not exposed to the network.
		runtimeConfig.GinHost = "0.0.0.0"
		runtimeConfig.GinPort = "9000"

		debugMode := os.Getenv("DEBUG_MODE")
		var err error = nil
		runtimeConfig.DebugMode, err = strconv.ParseBool(debugMode)
		if err != nil {
			runtimeConfig.DebugMode = true
		}

		runtimeConfig.PsqlHost = os.Getenv("PG_HOST")
		runtimeConfig.PsqlPort = os.Getenv("PG_PORT")
		runtimeConfig.PsqlDatabase = os.Getenv("PG_DATABASE")
		runtimeConfig.PsqlPassword = os.Getenv("PG_PASSWORD")
		runtimeConfig.PsqlUser = os.Getenv("PG_USER")
	}

	// Set Gin running mode based on value of the debug mode
	switch runtimeConfig.DebugMode {
	case true:
		gin.SetMode(gin.DebugMode)
	case false:
		gin.SetMode(gin.ReleaseMode)
	}

	// Instantiate a new stored... something
	// Although in our case it should only ever be a db.postgres instance
	c, err := db.NewRepository(fmt.Sprintf("postgres://%v:%v@%v:%v/%v",
		runtimeConfig.PsqlUser,
		runtimeConfig.PsqlPassword,
		runtimeConfig.PsqlHost,
		runtimeConfig.PsqlPort,
		runtimeConfig.PsqlDatabase))
	if err != nil {
		fmt.Printf("error connecting to datastore: %s\n", err)
		return 8
	}
	defer c.Store.Disconnect()

	c.Store.Ping(context.Background())

	// Define the Gin router
	router := gin.Default()

	// Set up endpoints
	endpoints.Configure(router, &c)

	// Start the router
	err = router.Run(fmt.Sprintf("%v:%v", runtimeConfig.GinHost, runtimeConfig.GinPort))
	if err != nil {
		fmt.Printf("error running gin: %s\n", err)
		return 2
	}

	return 0
}
