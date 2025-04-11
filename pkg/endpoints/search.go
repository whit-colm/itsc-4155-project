package endpoints

import (
	"encoding/json"
	"net/http"
	"slices"
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
		domains []string
		query   string // TODO: "[]S"?
		limit   int
		offset  int
	)
	domains = strings.Split(c.Param("idx"), ",")
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

	// Make slices for each search domain
	if slices.Contains(domains, "comments") {
		// Search comments
	}
	if slices.Contains(domains, "booktitle") {
		// Search book titles
	}
	if slices.Contains(domains, "authorname") {
		// Search author names
	}

	ret := make([]*any, 25)
	for range limit {
		// Effectively do the merge part of merge sort
		// Because each slice will already be sorted by best scoring
		// first, we simply pick which top value of the search domains
		// has the highest out of all of them, add it to the `ret`
		// slice, then pop said slice.
	}
	retJson, err := json.Marshal(ret)
	if err != nil {
		panic(err) // TODO: Not this
	}
	c.JSON(http.StatusOK, retJson)
}
