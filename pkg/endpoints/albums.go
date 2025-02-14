package endpoints

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whit-colm/itsc-4155-project/pkg/models"
)

func GetAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, models.Albums)
}
