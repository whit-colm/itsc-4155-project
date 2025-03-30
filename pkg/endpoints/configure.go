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

	a := authHandle{}
	router.GET("/api/auth/login", a.Login)
	router.GET("/api/auth/github/callback", wrap(a.GithubCallback))
	router.GET("/api/auth/logout", wrap(a.Logout))

	u := userHandle{}
	router.GET("/user/:id", u.UserInfo)
	router.GET("/user/me", u.AccountInfo) // Only to be used by authenticated accts
	router.DELETE("/user/me", u.Delete)   // Only to be used by authenticated accts

	b := bookHandle{rp.Book}
	router.GET("/api/books", b.GetBooks)
	router.POST("/api/books/new", b.AddBook)
	router.GET("/api/books/:id", b.GetBookByID)
	router.GET("/api/books/isbn/:isbn", b.GetBookByISBN)

	l := blobHandle{}
	router.GET("/api/blob/:id", l.GetDecoded)
	router.GET("/api/blob/:id/object", l.GetRaw)
	router.POST("/api/blob/new", l.New)             // Only to be used by site admin or system itself
	router.DELETE("/api/blob/delete/:id", l.Delete) // Only to be used by site admin or system itself
}
