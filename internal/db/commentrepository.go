package db

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type commentRepository struct {
	db *pgxpool.Pool
}

// Useful to check that a type implements an interface
var _ repository.CommentManager[string] = (*commentRepository)(nil)

func newCommentRepository(psql *postgres) repository.CommentManager[string] {
	return &commentRepository{db: psql.db}
}

// GetBookComments implements repository.CommentManager.
func (c *commentRepository) BookComments(ctx context.Context, bookID uuid.UUID) ([]*model.Comment, error) {
	const errorCaller string = "book comments"
	comments := []*model.Comment{}

	rows, err := c.db.Query(ctx,
		`SELECT 
			 c.id,
			 c.book_id,
			 c.body,
			 c.rating,
			 c.parent_comment_id,
			 c.votes,
			 c.deleted,
			 c.created_at,
			 c.updated_at,
			 u.id,
			 COALESCE(u.display_name, u.handle, 'Deleted'),
			 COALESCE(u.pronouns, ''),
			 COALESCE(u.handle, 'deleted user'),
			 COALESCE(u.discriminator, 0),
			 u.avatar
		 FROM comments c
		 LEFT JOIN users u ON c.poster_id = u.id
		 WHERE c.book_id = $1
		 GROUP BY c.id`,
		bookID,
	)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", errorCaller, err)
	}
	defer rows.Close()

	for rows.Next() {
		var co model.Comment
		var cu model.CommentUser

		var ed time.Time
		var hn string
		var de int16

		if err := rows.Scan(
			&co.ID, &co.Book, &co.Body, &co.Rating, &co.Parent,
			&co.Votes, &co.Deleted, &co.Date, &ed, &cu.DisplayName,
			&cu.Pronouns, &hn, &de, &cu.Avatar,
		); err != nil {
			return nil, fmt.Errorf("%v: %w", errorCaller, err)
		}
		if ed.After(co.Date.Add(5 * time.Minute)) {
			co.Edited = ed
		}
		co.Poster = cu

		comments = append(comments, &co)
	}

	return comments, rows.Err()
}

// Create implements repository.CommentManager.
func (c *commentRepository) Create(ctx context.Context, comment *model.Comment) error {
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
func (c *commentRepository) Delete(ctx context.Context, commentID uuid.UUID) error {
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
func (c *commentRepository) GetByID(ctx context.Context, commentID uuid.UUID) (*model.Comment, error) {
	const errorCaller string = "get comment"
	var co model.Comment
	var cu model.CommentUser

	var ed time.Time
	var hn string
	var de int16

	if err := c.db.QueryRow(ctx,
		`SELECT 
			 c.id,
			 c.book_id,
			 c.body,
			 c.rating,
			 c.parent_comment_id,
			 c.votes,
			 c.deleted,
			 c.created_at,
			 c.updated_at,
			 u.id,
			 COALESCE(u.display_name, u.handle, 'Deleted'),
			 COALESCE(u.pronouns, ''),
			 COALESCE(u.handle, 'deleted user'),
			 COALESCE(u.discriminator, 0),
			 u.avatar
		 FROM comments c
		 LEFT JOIN users u ON c.poster_id = u.id
		 WHERE c.book_id = $1
		 GROUP BY c.id`,
		commentID,
	).Scan(
		&co.ID, &co.Book, &co.Body, &co.Rating, &co.Parent,
		&co.Votes, &co.Deleted, &co.Date, &ed, &cu.DisplayName,
		&cu.Pronouns, &hn, &de, &cu.Avatar,
	); err != nil {
		return nil, fmt.Errorf("%v: %w", errorCaller, err)
	}

	if ed.After(co.Date.Add(5 * time.Minute)) {
		co.Edited = ed
	}
	co.Poster = cu

	return &co, nil
}

func (c *commentRepository) Search(ctx context.Context, terms ...string) ([]*model.Comment, error) {
	panic("unimplemented")
}

// Update implements repository.CommentManager.
func (c *commentRepository) Update(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
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
