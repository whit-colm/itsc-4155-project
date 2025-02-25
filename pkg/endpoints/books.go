package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/models"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type bookHandle struct {
	repo repository.BookManager
}

func (bh *bookHandle) GetBooks(c *gin.Context) {
	s, err := bh.repo.Search(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			jsonParsableError{Summary: "failed to retrieve books",
				Details: err})
		return
	}
	c.JSON(http.StatusOK, s)
}

func (bh *bookHandle) GetBookByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			jsonParsableError{Summary: "Unable to parse UUID",
				Details: err})
		return
	}

	s, err := bh.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound,
			jsonParsableError{Summary: "Could not find book with ID",
				Details: err})
		return
	}
	c.JSON(http.StatusOK, *s)
}

func (bh *bookHandle) AddBook(c *gin.Context) {
	jsonData, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusInternalServerError,
			jsonParsableError{Summary: "failed to parse request body as JSON",
				Details: err})
		return
	}
	var b models.Book

	if err = json.Unmarshal(jsonData, &b); err != nil {
		c.JSON(http.StatusInternalServerError,
			jsonParsableError{Summary: "failed to unmarshal JSON into book object",
				Details: err})
		return
	}

	if err = bh.repo.Create(c.Request.Context(), &b); err != nil {
		c.JSON(http.StatusInternalServerError,
			jsonParsableError{Summary: "failed to add book to datastore",
				Details: err})
		return
	}

	c.IndentedJSON(http.StatusOK, b)
}
