package endpoints

import (
	"fmt"
	"net/http"
	"net/url"
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
	domains = strings.Split(c.Query("d"), ",")
	if q, err := url.QueryUnescape(c.Query("q")); err != nil {
		return http.StatusBadRequest,
			"Could not parse your query as a URL-encoded string",
			fmt.Errorf("%v: %w", errorCaller, err)
	} else if q == "" {
		return http.StatusBadRequest,
			"Your query must not be empty",
			nil
	} else {
		query = q
	}
	if r, err := strconv.Atoi(c.Query("r")); err != nil {
		limit = 25
	} else if r > 250 {
		limit = 250
	} else {
		limit = r
	}
	if o, err := strconv.Atoi(c.Query("o")); err != nil {
		offset = 0
	} else {
		offset = o
	}

	results := [][]repository.AnyScoreItemer{}
	if slices.Contains(domains, "comments") {
		// TODO: find a way to do muilti-domain offsets without. this.
		_, comments, err := h.comm.Search(c.Request.Context(), 0, limit+offset, query)
		if err != nil {
			return http.StatusServiceUnavailable,
				errorCaller, err
		}
		results = append(results, comments)
	}
	if slices.Contains(domains, "booktitle") {
		// TODO: find a way to do muilti-domain offsets without. this.
		_, booktitle, err := h.book.Search(c.Request.Context(), 0, limit+offset, query)
		if err != nil {
			return http.StatusServiceUnavailable,
				errorCaller, err
		}
		results = append(results, booktitle)
	}
	if slices.Contains(domains, "authorname") {
		// TODO: find a way to do muilti-domain offsets without. this.
		_, authorname, err := h.athr.Search(c.Request.Context(), 0, limit+offset, query)
		if err != nil {
			return http.StatusServiceUnavailable,
				errorCaller, err
		}
		results = append(results, authorname)
	}
	if len(results) == 0 {
		return http.StatusNotFound,
			"Results was empty. Like Absolutely *Nothing* empty. Are you sure you provided valid domain(s)?",
			fmt.Errorf("%v: results `%#v` for domains `%v` -> `%s`",
				errorCaller,
				results,
				c.Query("d"),
				domains,
			)
	}

	ret := make([]any, limit+offset)
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
			if s := v[0].ScoreValue(); s > highestScore.Score {
				highestScore.Idx = i
				highestScore.Score = s
			}
		}
		// If the index of the highest score is -1 we know we have
		// exhausted all results and should cut our losses.
		if highestScore.Idx == -1 {
			break
		}

		ret[i] = results[highestScore.Idx][0].ItemAsAny()
		results[highestScore.Idx] = results[highestScore.Idx][1:]
	}
	// TODO: THIS IS A BAD BAD BAD BAD BAD WAY OF DEALING WITH OFFSETS
	ret = ret[offset:]

	c.JSON(http.StatusOK, ret)
	return http.StatusOK, "", nil
}
