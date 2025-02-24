package endpoints

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whit-colm/itsc-4155-project/pkg/db"
)

type BookHandler struct {
	repo db.BookRepositoryManager
}

func GetBooks(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, "{}")
}

func AddBooks(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, "{}")
}
