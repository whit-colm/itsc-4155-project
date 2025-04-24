package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type bookHandle[S comparable] struct {
	repo repository.BookManager[S]
}

func (bh *bookHandle[S]) GetBookByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest,
			jsonParsableError{Summary: "Unable to parse UUID",
				Details: err},
		)
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

// This is a bit hacky, but it's just a redirect to the UUID page.
func (bh *bookHandle[S]) GetBookByISBN(c *gin.Context) {
	isbn, err := model.NewISBN(c.Param("isbn"))
	if err != nil {
		c.JSON(http.StatusBadRequest,
			jsonParsableError{Summary: "Unable to parse ISBN",
				Details: err},
		)
		return
	}

	book, err := bh.repo.GetByISBN(c.Request.Context(), isbn)
	if err != nil {
		c.JSON(http.StatusNotFound,
			jsonParsableError{Summary: "Could not find book with ISBN",
				Details: err},
		)
		return
	}
	c.Redirect(http.StatusFound, fmt.Sprintf("/api/books/%s", book.ID))
}

func (bh *bookHandle[S]) AddBook(c *gin.Context) {
	jsonData, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest,
			jsonParsableError{Summary: "failed to parse request body as JSON",
				Details: err})
		return
	}
	var b model.Book

	if err = json.Unmarshal(jsonData, &b); err != nil {
		c.JSON(http.StatusBadRequest,
			jsonParsableError{Summary: "failed to unmarshal JSON into book object",
				Details: err})
		return
	}

	if id, err := uuid.NewV7(); err != nil {
		c.JSON(http.StatusInternalServerError,
			jsonParsableError{Summary: "failed to generate new UUID",
				Details: err})
		return
	} else {
		b.ID = id
	}

	if err = bh.repo.Create(c.Request.Context(), &b); err != nil {
		c.JSON(http.StatusInternalServerError,
			jsonParsableError{Summary: "failed to add book to datastore",
				Details: err})
		return
	}

	c.IndentedJSON(http.StatusCreated, b)
}
