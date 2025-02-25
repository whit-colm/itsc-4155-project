package endpoints

import (
	"github.com/gin-gonic/gin"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type jsonParsableError struct {
	Summary string `json:"summary"`
	Details error  `json:"details"`
}

// Configure all backend endpoints
func Configure(router *gin.Engine, rp *repository.Repository) {
	s := dataStore{rp.Store}
	router.GET("/api/health", s.Health)

	b := bookHandle{rp.Book}
	router.GET("/api/books", b.GetBooks)
	router.POST("/api/books/new", b.AddBook)
	router.GET("/api/books/:id", b.GetBookByID)
}
