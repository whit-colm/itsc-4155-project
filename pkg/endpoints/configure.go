package endpoints

import (
	"context"
	"encoding/json"

	"github.com/gin-gonic/gin"
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
func Configure(router *gin.Engine, rp *repository.Repository, c *oauth2.Config) {
	conf = c

	api := router.Group("/api")

	s := dataStore{rp.Store}
	api.GET("/health", s.Health)

	ah = authHandle{rp.User, rp.Auth}
	var err error

	jwtSigner.pub, jwtSigner.priv, err = rp.Auth.KeyPair(context.TODO())
	ah.auth.KeyPair(context.Background())
	if err != nil {
		panic(err)
	}
	api.GET("/auth/github/login", ah.Login)
	api.GET("/auth/github/callback", wrap(ah.GithubCallback))
	api.GET("/auth/logout", wrap(ah.Logout)).Use(AuthorizationJWT()) // Only to be used by authenticated accts

	profile := api.Group("/user")
	profile.Use(AuthorizationJWT())
	uh = userHandle{rp.User, rp.Blob}
	profile.GET("/:id", wrap(uh.UserInfo))
	profile.GET("/me", wrap(uh.AccountInfo)) // Only to be used by authenticated accts
	profile.PATCH("/me", uh.Update)          // Only to be used by authenticated accts
	profile.DELETE("/me", uh.Delete)         // Only to be used by authenticated accts

	bh = bookHandle{rp.Book}
	api.GET("/books", bh.GetBooks)
	api.POST("/books/new", bh.AddBook)
	api.GET("/books/:id", bh.GetBookByID)
	api.GET("/books/isbn/:isbn", bh.GetBookByISBN)

	blob := api.Group("/blob")
	lh = blobHandle{rp.Blob}
	blob.GET("/:id", lh.GetDecoded)
	blob.GET("/:id/object", lh.GetRaw)
	blob.POST("/new", lh.New)            // Only to be used by site admin or system itself
	blob.DELETE("/:id", wrap(lh.Delete)) // Only to be used by site admin or system itself
}
