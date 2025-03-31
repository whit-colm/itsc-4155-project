package endpoints

import (
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
			c.JSON(status, struct {
				Summary string `json:"summary"`
				Details string `json:"details"`
			}{
				Summary: summary,
				Details: err.Error(),
			})
			return
		}
	}
}

type jsonParsableError struct {
	Summary string `json:"summary"`
	Details error  `json:"details"`
}

var conf *oauth2.Config

// Configure all backend endpoints
func Configure(router *gin.Engine, rp *repository.Repository, c *oauth2.Config) {
	conf = c

	s := dataStore{rp.Store}
	router.GET("/api/health", s.Health)

	ah = authHandle{rp.User}
	router.GET("/api/auth/login", ah.Login)
	router.GET("/api/auth/github/callback", wrap(ah.GithubCallback))
	router.GET("/api/auth/logout", wrap(ah.Logout))

	uh = userHandle{rp.User, rp.Blob}
	router.GET("/user/:id", uh.UserInfo)
	router.GET("/user/me", uh.AccountInfo) // Only to be used by authenticated accts
	router.DELETE("/user/me", uh.Delete)   // Only to be used by authenticated accts

	bh = bookHandle{rp.Book}
	router.GET("/api/books", bh.GetBooks)
	router.POST("/api/books/new", bh.AddBook)
	router.GET("/api/books/:id", bh.GetBookByID)
	router.GET("/api/books/isbn/:isbn", bh.GetBookByISBN)

	lh = blobHandle{rp.Blob}
	router.GET("/api/blob/:id", lh.GetDecoded)
	router.GET("/api/blob/:id/object", lh.GetRaw)
	router.POST("/api/blob/new", lh.New)             // Only to be used by site admin or system itself
	router.DELETE("/api/blob/delete/:id", lh.Delete) // Only to be used by site admin or system itself
}
