package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	oauth2Endpoints "golang.org/x/oauth2/endpoints"

	"github.com/whit-colm/itsc-4155-project/internal/db"
	"github.com/whit-colm/itsc-4155-project/pkg/endpoints"
	"github.com/whit-colm/itsc-4155-project/pkg/scraper"
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

	OAuth2GithubClientID     string
	OAuth2GithubClientSecret string
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

	flag.StringVar(&runtimeConfig.OAuth2GithubClientID, "oa2ghclientid", "", "GitHub Application Client ID")
	flag.StringVar(&runtimeConfig.OAuth2GithubClientSecret, "oa2ghclientsecret", "", "GitHub Application Client Secret")

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

		runtimeConfig.OAuth2GithubClientID = os.Getenv("GH_CLIENTID")
		runtimeConfig.OAuth2GithubClientSecret = os.Getenv("GH_CLIENTSECRET")
	}

	// Set Gin running mode based on value of the debug mode
	switch runtimeConfig.DebugMode {
	case true:
		gin.SetMode(gin.DebugMode)
	case false:
		gin.SetMode(gin.ReleaseMode)
	}

	// Instantiate our concrete storage class (PostgreSQL)
	// Although as far as the rest of the program is concerned, it's a
	// bunch of repositories
	ds, err := db.NewRepository(
		fmt.Sprintf("postgres://%v:%v@%v:%v/%v",
			runtimeConfig.PsqlUser,
			runtimeConfig.PsqlPassword,
			runtimeConfig.PsqlHost,
			runtimeConfig.PsqlPort,
			runtimeConfig.PsqlDatabase,
		),
		30*time.Second,
	)
	if err != nil {
		fmt.Printf("error connecting to datastore: %s\n", err)
		return 8
	}
	defer ds.Store.Disconnect()

	ds.Store.Ping(context.Background())

	// Set up what we can of the OAuth2 config
	// the rest gets defined later, this is just how we pass secrets
	ghoa2 := oauth2.Config{
		ClientID:     runtimeConfig.OAuth2GithubClientID,
		ClientSecret: runtimeConfig.OAuth2GithubClientSecret,
		Scopes:       []string{"read:user", "user:email", "read:gpg_key"},
		Endpoint:     oauth2Endpoints.GitHub,
	}

	sc := scraper.NewBookScraper(ds.Blob, ds.Book, ds.Author)

	// Define the Gin router
	router := gin.Default()

	// Set up endpoints
	endpoints.Configure(router, &ds, &ghoa2, sc)

	// Start the router
	err = router.Run(fmt.Sprintf("%v:%v", runtimeConfig.GinHost, runtimeConfig.GinPort))
	if err != nil {
		fmt.Printf("error running gin: %s\n", err)
		return 2
	}

	return 0
}
