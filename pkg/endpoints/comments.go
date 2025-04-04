package endpoints

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type commentHandle struct {
	book repository.BookManager
	comm repository.CommentManager
}

var ch commentHandle

func (ch *commentHandle) BookReviews(c *gin.Context) (int, string, error) {
	const errorCaller string = "get book reviews"
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return http.StatusBadRequest,
			"Unable to parse UUID",
			fmt.Errorf("%s: %w", errorCaller, err)
	}
	comments, err := ch.comm.GetBookComments(c.Request.Context(), id)
	if err != nil {
		return wrapDatastoreError(errorCaller, err)
	}
	c.JSON(http.StatusOK, comments)
	return http.StatusOK, "", nil
}

func (ch *commentHandle) Get(c *gin.Context) (int, string, error) {
	const errorCaller string = "get comment"
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return http.StatusBadRequest,
			"Unable to parse UUID",
			fmt.Errorf("%s: %w", errorCaller, err)
	}
	comment, err := ch.comm.GetByID(c.Request.Context(), id)
	if err != nil {
		return wrapDatastoreError(errorCaller, err)
	}
	c.JSON(http.StatusOK, comment)
	return http.StatusOK, "", nil
}

func (ch *commentHandle) Post(c *gin.Context) (int, string, error) {
	const errorCaller string = "post new comment"
	// The user ID parameter must be set.
	userID, err := wrapGinContextUserID(c)
	if err != nil && !errors.Is(err, errUserIDKeyNotFound) {
		return http.StatusInternalServerError,
			"issue parsing ID from context",
			fmt.Errorf("%s: %w", errorCaller, err)
	} else if errors.Is(err, errUserIDKeyNotFound) {
		return http.StatusUnauthorized,
			"you must be logged in to access this page",
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
	comment.Poster.ID = userID
	if err = ch.comm.Create(c.Request.Context(), &comment); err != nil {
		return wrapDatastoreError(errorCaller, err)
	}

	return http.StatusOK, "", nil
}

func (ch *commentHandle) Edit(c *gin.Context) (int, string, error) {
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

	storedComment, err = ch.comm.Update(c.Request.Context(), newComment.ID, &newComment)
	if err != nil {
		return wrapDatastoreError(errorCaller, err)
	}
	c.JSON(http.StatusOK, storedComment)

	return http.StatusOK, "", nil
}

func (ch *commentHandle) Delete(c *gin.Context) (int, string, error) {
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

	authorID, err := ch.comm.GetAuthor(c.Request.Context(), commentIDParam)
	if err != nil {
		return wrapDatastoreError(errorCaller, err)
	} else if userIDParam != authorID && !perms {
		return http.StatusForbidden,
			"You do not have permission to delete that comment",
			nil
	}
	return http.StatusOK, "", nil
}
