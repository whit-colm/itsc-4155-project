package endpoints

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type searchHandle[S comparable] struct {
	book repository.BookManager[S]
	athr repository.AuthorManager[S]
	comm repository.CommentManager[S]
}

func (h searchHandle[S]) Search(c *gin.Context) {
	const errorCaller string = "search"
	// Do search query stuff

	var (
		idx    []string
		query  string // TODO: "[]S"?
		limit  int
		offset int
	)
	idx = strings.Split(c.Param("idx"), ",")
	query = c.Param("q")
	if r, err := strconv.Atoi(c.Param("r")); err != nil {
		limit = 25
	} else {
		limit = r
	}
	if o, err := strconv.Atoi(c.Param("o")); err != nil {
		offset = 0
	} else {
		offset = o
	}

	panic("unimplemented")
}
