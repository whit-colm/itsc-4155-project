package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
	"golang.org/x/oauth2"
)

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

	u := userHandle{}
	router.GET("/api/auth/login", u.Login)
	router.GET("/api/auth/github/callback", u.GithubCallback)
	router.GET("/api/auth/logout", u.Logout)
	router.GET("/user/me", u.AccountInfo)

	b := bookHandle{rp.Book}
	router.GET("/api/books", b.GetBooks)
	router.POST("/api/books/new", b.AddBook)
	router.GET("/api/books/:id", b.GetBookByID)
	router.GET("/api/books/isbn/:isbn", b.GetBookByISBN)
}
