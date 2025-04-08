package mockdatastore

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type VoteRepo struct {
	mut sync.RWMutex
	// commentID -> userID -> vote
	votes map[uuid.UUID]map[uuid.UUID]int8
}

var _ repository.VoteManager = (*VoteRepo)(nil)

// UserVotes implements repository.VoteManager.
func (m *VoteRepo) UserVotes(ctx context.Context, userID uuid.UUID) (map[uuid.UUID]int8, error) {
	m.mut.RLock()
	defer m.mut.RUnlock()
	result := make(map[uuid.UUID]int8)

	for cID, u := range m.votes {
		for uID, v := range u {
			if uID == userID {
				result[cID] = v
			}
		}
	}
	return result, nil
}

// Voted implements repository.VoteManager.
func (m *VoteRepo) Voted(ctx context.Context, userID uuid.UUID, commentIDs uuid.UUIDs) (map[uuid.UUID]int8, error) {
	m.mut.RLock()
	defer m.mut.RUnlock()
	result := make(map[uuid.UUID]int8)

	for _, cID := range commentIDs {
		result[cID] = m.votes[cID][userID]
	}

	return result, nil
}

func (m *VoteRepo) Vote(ctx context.Context, userID uuid.UUID, commentID uuid.UUID, vote int) (int, error) {
	m.mut.Lock()
	defer m.mut.Unlock()

	userVotes, exists := m.votes[commentID]
	if !exists {
		userVotes = make(map[uuid.UUID]int8)
		m.votes[commentID] = userVotes
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
