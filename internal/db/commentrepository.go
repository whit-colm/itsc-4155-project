package db

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type commentRepository[S comparable] struct {
	db *pgxpool.Pool
}

// Useful to check that a type implements an interface
var _ repository.CommentManager[string] = (*commentRepository[string])(nil)

func newCommentRepository(psql *postgres) repository.CommentManager[string] {
	return &commentRepository[string]{db: psql.db}
}

// queryString constructs a SELECT statement for retrieving comment records from the database.
// It optionally includes a search scoring column and uses the provided clause as a filtering condition.
//
// Parameters:
//
//	clause - The condition appended to the WHERE clause, affecting which comment rows are returned.
//	search - A boolean toggle indicating whether to include search scoring in the result.
//
// Returns:
//
//	A fully composed SQL query string for comment retrieval, with optional search scoring.
func (c commentRepository[S]) queryString(clause string, search bool) string {
	return fmt.Sprintf(`SELECT
			 %v
			 c.id,
			 c.book_id,
			 COALESCE(c.body, ''),
			 COALESCE(c.rating, -1.0),
			 c.parent_comment_id,
			 c.votes,
			 c.deleted,
			 c.created_at,
			 c.updated_at,
			 u.id,
			 COALESCE(u.display_name, u.handle, 'Deleted'),
			 COALESCE(u.pronouns, ''),
			 COALESCE(u.handle, 'deleted'),
			 COALESCE(u.discriminator, 0),
			 u.avatar
		 FROM comments c
		 LEFT JOIN users u ON c.poster_id = u.id
		 WHERE %v`,
		func() string {
			if search {
				return "paradedb.score(c.id),"
			}
			return ""
		}(),
		clause,
	)
}

// rowsParse retrieves and parses comment data from the provided pgx.Rows.
// If the search parameter is true, it expects an additional float32 score as
// the first scanned column. The function then scans data into a model.Comment
// along with associated user information, determines whether the comment
// should be flagged as edited based on creation and update timestamps, and
// constructs the comment's poster username using handle components. It returns
// the populated *model.Comment, a float32 representing the comment's search
// score, and an error if any field scanning or username construction fails.
func (c commentRepository[S]) rowsParse(rows pgx.Rows, search bool) (*model.Comment, float32, error) {
	var cmt model.Comment
	var cmtUser model.CommentUser
	var s float32

	var e time.Time
	var h string
	var d int16

	if search {
		if err := rows.Scan(
			&s, &cmt.ID, &cmt.Book, &cmt.Body, &cmt.Rating,
			&cmt.Parent, &cmt.Votes, &cmt.Deleted, &cmt.Date, &e,
			&cmtUser.ID, &cmtUser.DisplayName, &cmtUser.Pronouns, &h,
			&d, &cmtUser.Avatar,
		); err != nil {
			return nil, -1.0, err
		}
	} else {
		if err := rows.Scan(
			&cmt.ID, &cmt.Book, &cmt.Body, &cmt.Rating, &cmt.Parent,
			&cmt.Votes, &cmt.Deleted, &cmt.Date, &e, &cmtUser.ID,
			&cmtUser.DisplayName, &cmtUser.Pronouns, &h, &d, &cmtUser.Avatar,
		); err != nil {
			return nil, -1.0, err
		}
	}

	if e.After(cmt.Date.Add(5 * time.Minute)) {
		cmt.Edited = e
	}
	if uname, err := model.UsernameFromComponents(h, d); err != nil {
		return nil, -1.0, err
	} else {
		cmtUser.Username = uname
	}
	cmt.Poster = cmtUser

	return &cmt, s, nil
}

// GetBookComments implements repository.CommentManager.
func (c *commentRepository[S]) BookComments(ctx context.Context, bookID uuid.UUID) ([]*model.Comment, error) {
	const errorCaller string = "book comments"
	comments := []*model.Comment{}

	rows, err := c.db.Query(ctx,
		c.queryString("c.book_id = $1", false),
		bookID,
	)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", errorCaller, err)
	}
	defer rows.Close()

	for rows.Next() {
		cmt, _, err := c.rowsParse(rows, false)
		if err != nil {
			return nil, fmt.Errorf("%v: %w", errorCaller, err)
		}
		comments = append(comments, cmt)
	}

	return comments, rows.Err()
}

// Create implements repository.CommentManager.
func (c *commentRepository[S]) Create(ctx context.Context, comment *model.Comment) error {
	const errorCaller string = "create comment"
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %w", errorCaller, err)
	}
	defer tx.Rollback(ctx)

	now := time.Now()
	if comment.Parent == uuid.Nil {
		_, err = tx.Exec(ctx,
			`INSERT INTO comments (
				 id, book_id, poster_id, body, rating, created_at, 
				 updated_at
			 ) VALUES (
				 $1, $2, $3, $4, $5, $6, $7
			 )`,
			comment.ID, comment.Book, comment.Poster.ID, comment.Body,
			comment.Rating, now, now,
		)
	} else {
		_, err = tx.Exec(ctx,
			`INSERT INTO comments (
				 id, book_id, poster_id, body, parent_comment_id, 
				 created_at, updated_at
			 ) VALUES (
				 $1, $2, $3, $4, $5, $6, $7
			 )`,
			comment.ID, comment.Book, comment.Poster.ID, comment.Body,
			comment.Parent, now, now,
		)
	}
	if err != nil {
		return fmt.Errorf("%v: %w", errorCaller, err)
	}

	if _, err = tx.Exec(ctx,
		`INSERT INTO votes (comment_id, user_id, vote) VALUES ($2, $1, 1)`,
		comment.Poster.ID, comment.ID,
	); err != nil {
		return fmt.Errorf("%v: %w", errorCaller, err)
	}

	return tx.Commit(ctx)
}

// Delete implements repository.CommentManager.
func (c *commentRepository[S]) Delete(ctx context.Context, commentID uuid.UUID) error {
	const errorCaller string = "delete comment"
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %w", errorCaller, err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx,
		`DELETE FROM comments
		 WHERE id = $1`,
		commentID,
	); err != nil {
		return fmt.Errorf("%v: %w", errorCaller, err)
	}

	return tx.Commit(ctx)
}

// GetByID implements repository.CommentManager.
func (c *commentRepository[S]) GetByID(ctx context.Context, commentID uuid.UUID) (*model.Comment, error) {
	const errorCaller string = "get comment"
	var co model.Comment
	r, err := c.db.Query(ctx,
		c.queryString("c.id = $1", false),
		commentID,
	)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", errorCaller, err)
	}
	defer r.Close()

	// We should only get one row back
	multiple := false
	for r.Next() {
		if multiple {
			return nil, fmt.Errorf("%v: multiple rows returned", errorCaller)
		}
		multiple = true

		cmt, _, err := c.rowsParse(r, false)
		if err != nil {
			return nil, fmt.Errorf("%v: %w", errorCaller, err)
		}
		co = *cmt
	}

	return &co, r.Err()
}

// Search implements repository.CommentManager.
func (c *commentRepository[S]) Search(ctx context.Context, offset int, limit int, query ...string) ([]repository.SearchResult[model.Comment], []repository.AnyScoreItemer, error) {
	const errorCaller string = "comment search"
	var resultsT []repository.SearchResult[model.Comment]
	var resultsASI []repository.AnyScoreItemer

	qStr := strings.Join(query, " ")
	rows, err := c.db.Query(ctx,
		c.queryString(`c.body @@@ $1
			 ORDER BY paradedb.score(c.id) DESC, updated_at DESC
			 LIMIT $2 OFFSET $3`,
			true,
		),
		qStr,
		limit,
		offset,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("%v: %w", errorCaller, err)
	}

	for rows.Next() {
		c, s, err := c.rowsParse(rows, true)
		if err != nil {
			return nil, nil, fmt.Errorf("%v: %w", errorCaller, err)
		}

		r := repository.SearchResult[model.Comment]{
			Item:  c,
			Score: s,
		}
		resultsT = append(resultsT, r)
		resultsASI = append(resultsASI, r)
	}

	return resultsT, resultsASI, rows.Err()
}

// Update implements repository.CommentManager.
func (c *commentRepository[S]) Update(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	const errorCaller string = "update comment"
	cc, err := c.GetByID(ctx, comment.ID)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", errorCaller, err)
	}
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", errorCaller, err)
	}
	defer tx.Rollback(ctx)

	// Check to make sure none of the unalterable values are changed
	if cc.Book != comment.Book ||
		cc.Poster.ID != comment.Poster.ID ||
		cc.Parent != comment.ID ||
		cc.Date != comment.Date {
		return nil, fmt.Errorf("%v: attempting to update immutable fields", errorCaller)
	}
	// Clobber the edited field
	comment.Edited = time.Now()

	if _, err = tx.Exec(ctx,
		`UPDATE users SET (
			 body,
			 updated_at,
		 ) = (
			 $2, $3,
		 ) WHERE id=$1`,
		comment.ID, comment.Body, comment.Edited,
	); err != nil {
		return nil, fmt.Errorf("%v: %w", errorCaller, err)
	}

	if comment.Parent == uuid.Nil {
		// If the parent is nil, then the rating can be updated without
		// issue
		if _, err = tx.Exec(ctx,
			`UPDATE users SET (
				 rating,
			 ) = (
				 $2,
			 ) WHERE id=$1`,
			comment.ID, comment.Rating,
		); err != nil {
			return nil, fmt.Errorf("%v: %w", errorCaller, err)
		}
	}

	return comment, tx.Commit(ctx)
}
