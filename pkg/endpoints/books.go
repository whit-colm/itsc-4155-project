package endpoints

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whit-colm/itsc-4155-project/pkg/models"
)

func GetBooks(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, models.Books)
}

func AddBooks(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, models.Books)
}
