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

func (h searchHandle[S]) Search(c *gin.Context) (int, string, error) {
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

	results := [][]repository.AnyScoreItemer{}
	if slices.Contains(domains, "comments") {
		comments, err := h.comm.Search(c.Request.Context(), offset, limit, query)
		if err != nil {
			return http.StatusServiceUnavailable,
				errorCaller, err
		}
		asi := make([]repository.AnyScoreItemer, len(comments))
		for _, v := range comments {
			asi = append(asi, v)
		}
		results = append(results, asi)
	}
	if slices.Contains(domains, "booktitle") {
		booktitle, err := h.book.Search(c.Request.Context(), offset, limit, query)
		if err != nil {
			return http.StatusServiceUnavailable,
				errorCaller, err
		}
		asi := make([]repository.AnyScoreItemer, len(booktitle))
		for _, v := range booktitle {
			asi = append(asi, v)
		}
		results = append(results, asi)
	}
	if slices.Contains(domains, "authorname") {
		authorname, err := h.athr.Search(c.Request.Context(), offset, limit, query)
		if err != nil {
			return http.StatusServiceUnavailable,
				errorCaller, err
		}
		asi := make([]repository.AnyScoreItemer, len(authorname))
		for _, v := range authorname {
			asi = append(asi, v)
		}
		results = append(results, asi)
	}

	ret := make([]any, limit)
	for i := range ret {
		// Effectively do the merge part of merge sort
		// Because each slice will already be sorted by best scoring
		// first, we simply pick which top value of the search domains
		// has the highest out of all of them, add it to the `ret`
		// slice, then pop said slice.
		highestScore := struct {
			Idx   int
			Score float64
		}{
			-1, 0.0,
		}
		for i, v := range results {
			if len(v) < 1 {
				continue
			}
			if s := v[0].Score(); s > highestScore.Score {
				highestScore.Idx = i
				highestScore.Score = s
			}
		}
		ret[i] = results[highestScore.Idx][0].Item()
		results[highestScore.Idx] = results[highestScore.Idx][1:]
	}
	retJson, err := json.Marshal(ret)
	if err != nil {
		return http.StatusInternalServerError,
			errorCaller, err
	}
	c.JSON(http.StatusOK, retJson)
}
