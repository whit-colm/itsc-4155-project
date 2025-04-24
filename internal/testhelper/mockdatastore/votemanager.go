package mockdatastore

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type VoteRepo[S comparable] struct {
	mut  sync.RWMutex
	user repository.UserManager
	comm repository.CommentManager[S]
	// commentID -> userID -> vote
	votes map[uuid.UUID]map[uuid.UUID]int8
}

var _ repository.VoteManager = (*VoteRepo[string])(nil)

func (r *VoteRepo[S]) prune(ctx context.Context) {
	r.mut.Lock()
	defer r.mut.Unlock()
	var deadComments uuid.UUIDs
	for cID, uVotes := range r.votes {
		if _, err := r.comm.GetByID(ctx, cID); errors.Is(err, repository.ErrNotFound) {
			deadComments = append(deadComments, cID)
			continue
		} else if err != nil {
			panic(err)
		}
		var deadUsers uuid.UUIDs
		for uID := range uVotes {
			if _, err := r.user.GetByID(ctx, uID); errors.Is(err, repository.ErrNotFound) {
				deadUsers = append(deadUsers, uID)
			} else if err != nil {
				panic(err)
			}
		}
		for _, uID := range deadUsers {
			delete(uVotes, uID)
		}
	}
	for _, cID := range deadComments {
		delete(r.votes, cID)
	}
}

// UserVotes implements repository.VoteManager.
func (r *VoteRepo[S]) UserVotes(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]int8, error) {
	r.prune(ctx)
	r.mut.RLock()
	defer r.mut.RUnlock()
	result := make(map[uuid.UUID]int8)

	for cID, u := range r.votes {
		for uID, v := range u {
			if uID == userID {
				result[cID] = v
			}
		}
	}
	return result, nil
}

// Voted implements repository.VoteManager.
func (r *VoteRepo[S]) Voted(ctx context.Context, userID uuid.UUID, commentIDs uuid.UUIDs) (map[uuid.UUID]int8, error) {
	r.prune(ctx)
	r.mut.RLock()
	defer r.mut.RUnlock()
	result := make(map[uuid.UUID]int8)

	for _, cID := range commentIDs {
		result[cID] = r.votes[cID][userID]
	}

	return result, nil
}

func (r *VoteRepo[S]) Vote(ctx context.Context, userID uuid.UUID, commentID uuid.UUID, vote int) (int, error) {
	r.prune(ctx)
	r.mut.Lock()
	defer r.mut.Unlock()

	userVotes, exists := r.votes[commentID]
	if !exists {
		userVotes = make(map[uuid.UUID]int8)
		r.votes[commentID] = userVotes
	}

	switch {
	case vote == 0:
		delete(userVotes, userID)
	case vote > 0:
		userVotes[userID] = 1
	case vote < 0:
		userVotes[userID] = -1
	}

	total := 0
	for _, v := range userVotes {
		total += int(v)
	}
	return total, nil
}
