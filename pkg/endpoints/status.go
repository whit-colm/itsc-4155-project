package endpoints

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type dataStore struct {
	store repository.StoreManager
}

func (ds *dataStore) Health(c *gin.Context) {
	if err := ds.store.Ping(c.Request.Context()); err != nil {
		c.String(http.StatusBadGateway, err.Error())
		return
	}
	c.String(http.StatusOK, "ok")
}
