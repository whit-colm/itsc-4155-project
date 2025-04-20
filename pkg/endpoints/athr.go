package endpoints

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type athrHandle[S comparable] struct {
	repo repository.AuthorManager[S]
}

var th athrHandle[string]

func (ah *athrHandle[S]) GetAuthorByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest,
			jsonParsableError{Summary: "Unable to parse UUID",
				Details: err},
		)
		return
	}

	s, err := ah.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound,
			jsonParsableError{Summary: "Could not find author with ID",
				Details: err})
		return
	}
	c.JSON(http.StatusOK, *s)
}
