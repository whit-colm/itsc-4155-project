package endpoints

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
	"golang.org/x/oauth2"
)

// Wrap handlers to allow for cleaner code in more intensive endpoints
func wrap(ep func(*gin.Context) (int, string, error)) func(*gin.Context) {
	return func(c *gin.Context) {
		status, summary, err := ep(c)
		if err != nil {
			c.Error(err)
		}
		if status >= 400 {
			// Dumb-ass function for unwrapping an error
			// because sometimes it just! doesn't render!
			c.JSON(status, jsonParsableError{
				Summary: summary,
				Details: err,
			})
			return
		}
	}
}

func wrapDatastoreError(caller string, err error) (int, string, error) {
	if errors.Is(err, repository.ErrNotFound) {
		return http.StatusNotFound,
			"Could not find resource matching given key or description",
			fmt.Errorf("%v: %w", caller, err)
	} else if errors.Is(err, repository.ErrBadConnection) {
		return http.StatusServiceUnavailable,
			"There was an issue connecting to the datastore",
			fmt.Errorf("%v: %w", caller, err)
	} else if errors.Is(err, repository.ErrBadTypecast) {
		return http.StatusBadRequest,
			"Could not cast given value as necessary type",
			fmt.Errorf("%v: %w", caller, err)
	} else {
		return http.StatusInternalServerError,
			"An issue occured and your request could not be completed",
			fmt.Errorf("%v: %w", caller, err)
	}
}

// In the same way you might use c.GetBool/GetInt/etc, but instead for
// a uuid.UUID. This also gives an error which can be fed back with a
// 400.
func wrapGetUUID(c *gin.Context, key string) (uuid.UUID, error) {
	idStr := c.Param(key)
	idStr = strings.TrimPrefix(idStr, "/")
	idStr = strings.TrimSuffix(idStr, "/")
	return uuid.Parse(idStr)
}

type jsonParsableError struct {
	Summary string `json:"summary"`
	Details error  `json:"details"`
}

func (j jsonParsableError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Summary string `json:"summary"`
		Details string `json:"details"`
	}{
		Summary: j.Summary,
		// I love Grust.
		Details: func(e error) string {
			switch e {
			case nil:
				return "<nil>"
			default:
				return e.Error()
			}
		}(j.Details),
	})
}

var conf *oauth2.Config

// Configure all backend endpoints
func Configure[S comparable](router *gin.Engine, rp *repository.Repository[S], c *oauth2.Config, scraper repository.BookScraper) {
	conf = c

	api := router.Group("/api")

	s := dataStore{rp.Store}
	api.GET("/health", s.Health)

	sh := searchHandle[S]{rp.Book, rp.Author, rp.Comment, scraper}
	api.GET("/search", wrap(sh.Search))

	ah = authHandle{rp.User, rp.Auth}
	var err error

	jwtSigner.pub, jwtSigner.priv, err = ah.auth.KeyPair(context.TODO())
	if err != nil {
		panic(err)
	}
	api.GET("/auth/github/login", ah.Login)
	api.GET("/auth/github/callback", wrap(ah.GithubCallback))

	th := athrHandle[S]{rp.Author}
	api.GET("/authors/:id", th.GetAuthorByID)

	profile := api.Group("/user")
	profile.Use(AuthorizationJWT())
	uh = userHandle{rp.User, rp.Blob}
	profile.GET("/:id", wrap(uh.UserInfo))
	//profile.DELETE("/:id", uh.Delete).Use(UserPermissions())
	profile.GET("/me", wrap(uh.UserInfo))            // Only to be used by authenticated accts
	profile.PATCH("/me", wrap(uh.Update))            // Only to be used by authenticated accts
	profile.PUT("/me/avatar", wrap(uh.UpdateAvatar)) // Only to be used by authenticated accts
	profile.DELETE("/me", wrap(uh.Delete))           // Only to be used by authenticated accts

	books := api.Group("/books")
	bh := bookHandle[S]{rp.Book}
	books.POST("/new", bh.AddBook).Use(AuthorizationJWT(), UserPermissions())
	books.GET("/:id", bh.GetBookByID)
	books.GET("/isbn/:isbn", bh.GetBookByISBN)
	// See below for additional book endpoints

	comments := api.Group("/comments")
	comments.Use(AuthorizationJWT())
	ch := commentHandle[S]{rp.Book, rp.Comment, rp.Vote}
	books.GET("/:id/reviews", wrap(ch.BookReviews))
	books.GET("/:id/reviews/votes", wrap(ch.Votes)) // Only to be used by authenticated accts
	books.POST("/:id/reviews", wrap(ch.Post))       // Only to be used by authenticated accts
	comments.POST("/", wrap(ch.Post))               // Only to be used by authenticated accts
	comments.GET("/:id", wrap(ch.Get))
	comments.POST("/:id/vote", wrap(ch.Vote))                                          // Only to be used by authenticated accts
	comments.GET("/:id/vote", wrap(ch.Voted))                                          // Only to be used by authenticated accts
	comments.PATCH("/:id", wrap(ch.Edit))                                              // Only to be used by authenticated accts
	comments.DELETE(":id", wrap(ch.Delete)).Use(AuthorizationJWT(), UserPermissions()) // Only to be used by authenticated accts (+admin functionality)

	blob := api.Group("/blob")
	lh = blobHandle{rp.Blob}
	blob.GET("/:id", wrap(lh.GetRaw))
	blob.POST("/new", wrap(lh.New)).Use(AuthorizationJWT(), UserPermissions())      // Only to be used by site admins or system itself
	blob.DELETE("/:id", wrap(lh.Delete)).Use(AuthorizationJWT(), UserPermissions()) // Only to be used by site admins or system itself
}
