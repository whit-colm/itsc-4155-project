package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type commentHandle[S comparable] struct {
	book repository.BookManager[S]
	comm repository.CommentManager[S]
	vote repository.VoteManager
}

// TODO: This is not where I want to concrete this...
//
// Yes I know it was a Very dirty yucky hack in the first place
// var ch commentHandle[comparable]

func (ch *commentHandle[S]) BookReviews(c *gin.Context) (int, string, error) {
	const errorCaller string = "get book reviews"
	bookID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return http.StatusBadRequest,
			"Unable to parse UUID",
			fmt.Errorf("%s: %w", errorCaller, err)
	}
	comments, err := ch.comm.BookComments(c.Request.Context(), bookID)
	if err != nil {
		return wrapDatastoreError(errorCaller, err)
	}
	c.JSON(http.StatusOK, comments)
	return http.StatusOK, "", nil
}

func (ch *commentHandle[S]) Get(c *gin.Context) (int, string, error) {
	const errorCaller string = "get comment"
	commentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return http.StatusBadRequest,
			"Unable to parse UUID",
			fmt.Errorf("%s: %w", errorCaller, err)
	}
	comment, err := ch.comm.GetByID(c.Request.Context(), commentID)
	if err != nil {
		return wrapDatastoreError(errorCaller, err)
	}
	c.JSON(http.StatusOK, comment)
	return http.StatusOK, "", nil
}

func (ch *commentHandle[S]) Post(c *gin.Context) (int, string, error) {
	const errorCaller string = "post new comment"
	// The user ID parameter must be set.
	tokenUserID, err := wrapGinContextUserID(c)
	if errors.Is(err, errUserIDKeyNotFound) {
		return http.StatusUnauthorized,
			"you must be logged in to access this page",
			fmt.Errorf("%s: %w", errorCaller, err)
	} else if err != nil {
		return http.StatusInternalServerError,
			"issue parsing ID from context",
			fmt.Errorf("%s: %w", errorCaller, err)
	}
	// Try to get very likely non-existent ID
	bookID := func(c *gin.Context) uuid.UUID {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil {
			return uuid.Nil
		}
		_, err = ch.book.GetByID(c.Request.Context(), id)
		if errors.Is(err, repository.ErrorNotFound) {
			return uuid.Nil
		}
		return id
	}(c)

	jBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return http.StatusBadRequest,
			"There was an issue reading the body of your request",
			fmt.Errorf("%s: %w", errorCaller, err)
	}
	var comment model.Comment
	if err := json.Unmarshal(jBytes, &comment); err != nil {
		return http.StatusBadRequest,
			"could not parse JSON into comment object",
			fmt.Errorf("%s: %w", errorCaller, err)
	}
	if bookID != uuid.Nil {
		comment.Book = bookID
	}
	// We do not need to populate the rest, this should be enough for
	// the backing store
	comment.Poster.ID = tokenUserID
	if err = ch.comm.Create(c.Request.Context(), &comment); err != nil {
		return wrapDatastoreError(errorCaller, err)
	}

	return http.StatusOK, "", nil
}

func (ch *commentHandle[S]) Edit(c *gin.Context) (int, string, error) {
	const errorCaller string = "edit comment"
	// A request must be authenticated to access this page.
	userIDParam, err := wrapGinContextUserID(c)
	if err != nil && !errors.Is(err, errUserIDKeyNotFound) {
		return http.StatusInternalServerError,
			"issue parsing ID from context",
			fmt.Errorf("%s: %w", errorCaller, err)
	} else if errors.Is(err, errUserIDKeyNotFound) {
		return http.StatusUnauthorized,
			"you must be logged in to access this page",
			fmt.Errorf("%s: %w", errorCaller, err)
	}
	// The comment ID URL parameter must be set.
	commentIDParam, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return http.StatusBadRequest,
			"Unable to parse UUID",
			fmt.Errorf("%s: %w", errorCaller, err)
	}

	// Read body (JSON) into actual object
	jBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return http.StatusBadRequest,
			"There was an issue reading the body of your request",
			fmt.Errorf("%s: %w", errorCaller, err)
	}
	var newComment model.Comment
	if err := json.Unmarshal(jBytes, &newComment); err != nil {
		return http.StatusBadRequest,
			"could not parse JSON into comment object",
			fmt.Errorf("%s: %w", errorCaller, err)
	} else if newComment.Deleted {
		return http.StatusBadRequest,
			"You cannot delete a comment by editing it. use DELETE instead",
			fmt.Errorf("%s: attempting to change delete value", errorCaller)
	}

	// Comment ID processing and tests:
	//  1. The comment ID in the body should match the comment ID in the
	//	   URL parameters
	//  2. If they do not, then the ID in the body must be nil.
	if newComment.ID == uuid.Nil {
		newComment.ID = commentIDParam
	} else if newComment.ID != commentIDParam {
		return http.StatusForbidden,
			"Editing of other users' comments is not allowed",
			fmt.Errorf(
				"%s: mismatch between user IDs of comment `%v`, and authenticated user `%v`",
				errorCaller,
				newComment.Poster.ID,
				commentIDParam,
			)
	}

	// User ID processing and tests:
	//  1. The user ID in the comment body should match the user ID of
	// 	   the user associated with the authorization token
	//  2. If they do not, then the ID in the comment body must be nil.
	if newComment.Poster.ID == uuid.Nil {
		newComment.Poster.ID = userIDParam
	} else if newComment.Poster.ID != userIDParam {
		return http.StatusForbidden,
			"Editing of other users' comments is not allowed",
			fmt.Errorf(
				"%s: mismatch between user IDs of comment `%v`, and authenticated user `%v`",
				errorCaller,
				newComment.Poster.ID,
				commentIDParam,
			)
	}

	// Stored comment/New comment tests:
	//  1. The new comment cannot be deleted
	//  2. The old comment cannot be deleted
	//  3. The user IDs of the old and new comment must match
	storedComment, err := ch.comm.GetByID(c.Request.Context(), newComment.ID)
	if err != nil {
		return wrapDatastoreError(errorCaller, err)
	} else if storedComment.Deleted {
		return http.StatusGone,
			"This comment has been deleted and cannot be edited",
			nil
	} else if storedComment.Poster.ID != newComment.Poster.ID {
		return http.StatusForbidden,
			"Editing of other users' comments is not allowed",
			fmt.Errorf(
				"%s: mismatch between user IDs of comment `%v`, and authenticated user `%v`",
				errorCaller,
				storedComment.Poster.ID,
				newComment.Poster.ID,
			)
	}

	storedComment, err = ch.comm.Update(c.Request.Context(), &newComment)
	if err != nil {
		return wrapDatastoreError(errorCaller, err)
	}
	c.JSON(http.StatusOK, storedComment)

	return http.StatusOK, "", nil
}

func (ch *commentHandle[S]) Delete(c *gin.Context) (int, string, error) {
	const errorCaller string = "delete comment"
	// A request must be authenticated to access this page.
	userIDParam, err := wrapGinContextUserID(c)
	if err != nil && !errors.Is(err, errUserIDKeyNotFound) {
		return http.StatusInternalServerError,
			"issue parsing ID from context",
			fmt.Errorf("%s: %w", errorCaller, err)
	} else if errors.Is(err, errUserIDKeyNotFound) {
		return http.StatusUnauthorized,
			"you must be logged in to access this page",
			fmt.Errorf("%s: %w", errorCaller, err)
	}
	// Get permission level of authenticated user, only the author of a
	// comment and site administrators can delete a comment
	perms := c.GetBool("permissions")

	// The comment ID URL parameter must be set.
	commentIDParam, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return http.StatusBadRequest,
			"Unable to parse UUID",
			fmt.Errorf("%s: %w", errorCaller, err)
	}

	comment, err := ch.comm.GetByID(c.Request.Context(), commentIDParam)
	if err != nil {
		return wrapDatastoreError(errorCaller, err)
	} else if userIDParam != comment.Poster.ID && !perms {
		return http.StatusForbidden,
			"You do not have permission to delete that comment",
			nil
	}
	return http.StatusOK, "", nil
}

func (ch *commentHandle[S]) Vote(c *gin.Context) (int, string, error) {
	const errorCaller string = "vote on comment"
	// Get User ID (required) from middleware
	userID, err := wrapGinContextUserID(c)
	if errors.Is(err, errUserIDKeyNotFound) {
		return http.StatusUnauthorized,
			"you must be logged in to access this page",
			fmt.Errorf("%s: %w", errorCaller, err)
	} else if err != nil {
		return http.StatusInternalServerError,
			"issue parsing ID from context",
			fmt.Errorf("%s: %w", errorCaller, err)
	}
	// Get comment ID (required) from parameters
	commentID, err := wrapGetUUID(c, "id")
	if err != nil {
		return http.StatusBadRequest,
			"issue parsing ID from URL",
			fmt.Errorf("%v: %w", errorCaller, err)
	}
	// While this is a method used in a POST call, because the data is
	// so simple we just transmit it via a query parameter
	vote, err := strconv.Atoi(c.Query("vote"))
	if err != nil {
		return http.StatusBadRequest,
			"There was an issue processing the vote query parameter",
			fmt.Errorf("%v: %w", errorCaller, err)
	}

	total, err := ch.vote.Vote(c.Request.Context(), userID, commentID, vote)
	if err != nil {
		return wrapDatastoreError(errorCaller, err)
	}

	m := make(map[uuid.UUID]int)
	m[commentID] = total
	c.JSON(http.StatusOK, m)

	return http.StatusOK, "", nil
}

func (ch *commentHandle[S]) Voted(c *gin.Context) (int, string, error) {
	const errorCaller string = "get comment vote"
	// Get User ID (required) from middleware
	userID, err := wrapGinContextUserID(c)
	if errors.Is(err, errUserIDKeyNotFound) {
		return http.StatusUnauthorized,
			"you must be logged in to access this page",
			fmt.Errorf("%s: %w", errorCaller, err)
	} else if err != nil {
		return http.StatusInternalServerError,
			"issue parsing ID from context",
			fmt.Errorf("%s: %w", errorCaller, err)
	}
	// Get comment ID (required) from parameters
	commentID, err := wrapGetUUID(c, "id")
	if err != nil {
		return http.StatusBadRequest,
			"issue parsing ID from URL",
			fmt.Errorf("%v: %w", errorCaller, err)
	}

	vMap, err := ch.vote.Voted(c.Request.Context(), userID, uuid.UUIDs{commentID})
	if err != nil {
		return wrapDatastoreError(errorCaller, err)
	}

	// TODO: Talk with frontend how they want the value given

	// 1. same as `Votes`:
	c.JSON(http.StatusOK, vMap)

	// 2. Plain ole number
	//c.String(http.StatusOK, "%d", vMap[commentID])

	return http.StatusOK, "", nil
}

func (ch *commentHandle[S]) Votes(c *gin.Context) (int, string, error) {
	const errorCaller string = "get all votes on book"
	// Get User ID (required) from middleware
	userID, err := wrapGinContextUserID(c)
	if errors.Is(err, errUserIDKeyNotFound) {
		return http.StatusUnauthorized,
			"you must be logged in to access this page",
			fmt.Errorf("%s: %w", errorCaller, err)
	} else if err != nil {
		return http.StatusInternalServerError,
			"issue parsing ID from context",
			fmt.Errorf("%s: %w", errorCaller, err)
	}
	// Get Book ID (required) from parameters
	bookID, err := wrapGetUUID(c, "id")
	if err != nil {
		return http.StatusBadRequest,
			"issue parsing ID from URL",
			fmt.Errorf("%v: %w", errorCaller, err)
	}

	// TODO: THIS IS SO UNIMAGINABLY LABOR-INTENSIVE.
	// COME UP WITH A WAY TO NOT. DO THIS.
	ids, h, s, err := func(bookID uuid.UUID) (uuid.UUIDs, int, string, error) {
		comments, err := ch.comm.BookComments(c.Request.Context(), bookID)
		if err != nil {
			h, s, e := wrapDatastoreError(errorCaller, err)
			return nil, h, s, e
		}

		cIDs := make(uuid.UUIDs, len(comments))
		for i, v := range comments {
			cIDs[i] = v.ID
		}
		return cIDs, 0, "", nil
	}(bookID)
	if err != nil {
		return h, s, err
	}

	if votes, err := ch.vote.Voted(c.Request.Context(), userID, ids); err != nil {
		return wrapDatastoreError(errorCaller, err)
	} else {
		c.JSON(http.StatusOK, votes)
	}
	return http.StatusOK, "", nil
}
