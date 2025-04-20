package endpoints

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

// Is not strictly necessary for security reasons, but prevents queries
// from randomly failing due to unexpected characters.
var querySanitizer *regexp.Regexp

func init() {
	querySanitizer = regexp.MustCompile(`([\\\'\"\;\(\)\[\]\{\}\^\%\x60])`)
}

type searchHandle[S comparable] struct {
	book repository.BookManager[S]
	athr repository.AuthorManager[S]
	comm repository.CommentManager[S]
	scrp repository.BookScraper
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
		query = querySanitizer.ReplaceAllString(q, "\\$1")
	}
	if r, err := strconv.Atoi(c.Query("r")); err != nil {
		limit = 25
	} else {
		limit = min(r, 250)
	}
	if o, err := strconv.Atoi(c.Query("o")); err != nil {
		offset = 0
	} else {
		offset = o
	}

	results := [][]repository.AnyScoreItemer{}
	if slices.Contains(domains, "comments") {
		// TODO: find a way to do multi-domain offsets without. this.
		_, comments, err := h.comm.Search(c.Request.Context(), 0, limit+offset, query)
		if err != nil {
			return http.StatusServiceUnavailable,
				errorCaller, err
		}
		results = append(results, comments)
	}
	if slices.Contains(domains, "booktitle") {
		// Searching for books is special, because this is also the
		// method by which we scrape book data from our external
		// sources. If we do not find a sufficient number of results
		// we will make the same query to the scraper and retry.
		var booktitle []repository.AnyScoreItemer
		var err error
		for i := range 2 {
			if i == 1 {
				// If we are on the second iteration, we will scrape
				// and scrape until we have enough results
				// or we run out of results to scrape.
				added := 0
				iter := 0
				for added < limit {
					o := offset + iter*limit
					n, err := h.scrp.Scrape(c.Request.Context(), o, limit, query)
					if err != nil {
						break
					}
					if n == -1 {
						// If we got -1, we know there's nothing left to scrape
						break
					} else {
						added += n
						iter++
					}
				}
			}
			// TODO: find a way to do multi-domain offsets without. this.
			_, booktitle, err = h.book.Search(c.Request.Context(), 0, limit+offset, query)
			if err != nil {
				return http.StatusServiceUnavailable,
					errorCaller, err
			}
			// We don't need to scrape if we have enough results
			if len(booktitle) > limit+offset {
				break
			}
		}
		results = append(results, booktitle)
	}
	if slices.Contains(domains, "authorname") {
		// TODO: find a way to do multi-domain offsets without. this.
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

	processed := make([]map[string]any, limit+offset)
	for i := range processed {
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

		// Warning: This is unwell. We have to marshal the any to a byte array,
		// unmarshal it to a map, tack on the APIVersion, then re-marshal it
		// and *then* we can append it to the processed slice.
		item := results[highestScore.Idx][0]
		anyVal := item.ItemAsAny()

		// Marshal to JSON
		b, err := json.Marshal(anyVal)
		if err != nil {
			// Handle error, skip this item
			continue
		}

		// Unmarshal to map
		var m map[string]any
		if err := json.Unmarshal(b, &m); err != nil {
			// Handle error, skip this item
			continue
		}

		// Add APIVersion if possible
		if a := item.APIVersion(); a != "" {
			m["apiVersion"] = a
		}

		processed[i] = m
		results[highestScore.Idx] = results[highestScore.Idx][1:]
	}
	// TODO: THIS IS A BAD BAD BAD BAD BAD WAY OF DEALING WITH OFFSETS
	processed = processed[offset:]

	c.JSON(http.StatusOK, processed)
	return http.StatusOK, "", nil
}
