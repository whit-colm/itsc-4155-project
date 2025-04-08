package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type voteRepository struct {
	db *pgxpool.Pool
}

// Useful to check that a type implements an interface
var _ repository.VoteManager = (*voteRepository)(nil)

func newVoteRepository(psql *postgres) repository.VoteManager {
	return &voteRepository{db: psql.db}
}

// UserVotes implements repository.VoteManager.
func (r *voteRepository) UserVotes(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]int8, error) {
	const errorCaller string = "user votes"
	result := make(map[uuid.UUID]int8)
	rows, err := r.db.Query(ctx,
		`SELECT comment_id, vote FROM votes
		 WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", errorCaller, err)
	}
	defer rows.Close()

	for rows.Next() {
		var c uuid.UUID
		// the value stored is a SMALLINT, 2 bytes
		var v int16
		if err := rows.Scan(&c, &v); err != nil {
			return nil, fmt.Errorf("%v: %w", errorCaller, err)
		}
		result[c] = int8(v)
	}

	return result, rows.Err()
}

// Vote implements repository.CommentManager.
func (r *voteRepository) Vote(ctx context.Context, userID uuid.UUID, commentID uuid.UUID, vote int) (int, error) {
	const errorCaller string = "cast vote"
	var totalVotes int
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return totalVotes, fmt.Errorf("%v: %w", errorCaller, err)
	}
	defer tx.Rollback(ctx)

	switch vote {
	case 0:
		_, err = tx.Exec(ctx,
			`DELETE FROM votes v
		 	 WHERE comment_id = $2
		 	 	 AND user_id = $1`,
			userID, commentID,
		)
	default:
		if vote > 0 {
			vote = 1
		} else {
			vote = -1
		}
		_, err = tx.Exec(ctx,
			`INSERT INTO votes (comment_id, user_id, vote)
			 VALUES ($2, $1, $3)
			 ON CONFLICT (comment_id, user_id)
			 DO UPDATE SET vote = $3`,
			userID, commentID, vote,
		)
	}
	if err != nil {
		return totalVotes, fmt.Errorf("%v: %w", errorCaller, err)
	}

	if err = tx.QueryRow(ctx,
		`SELECT votes FROM comments WHERE id = $1`,
		commentID,
	).Scan(&totalVotes); err != nil {
		return totalVotes, fmt.Errorf("%v: %w", errorCaller, err)
	}

	return totalVotes, tx.Commit(ctx)
}

// Voted implements repository.CommentManager.
func (r *voteRepository) Voted(ctx context.Context, userID uuid.UUID, commentIDs uuid.UUIDs) (map[uuid.UUID]int8, error) {
	const errorCaller string = "cast vote"
	result := make(map[uuid.UUID]int8, len(commentIDs))
	if len(commentIDs) == 0 {
		return result, nil
	}
	rows, err := r.db.Query(ctx,
		`SELECT comment_id, vote FROM votes
		 WHERE user_id = $1 AND comment_id = ANY($2)`,
		userID, commentIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", errorCaller, err)
	}
	defer rows.Close()

	existing := make(map[uuid.UUID]int8)
	for rows.Next() {
		var c uuid.UUID
		// the value stored is a SMALLINT, 2 bytes
		var v int16
		if err := rows.Scan(&c, &v); err != nil {
			return nil, fmt.Errorf("%v: %w", errorCaller, err)
		}
		existing[c] = int8(v)
	}

	for _, cid := range commentIDs {
		result[cid] = existing[cid]
	}

	return result, rows.Err()
}
