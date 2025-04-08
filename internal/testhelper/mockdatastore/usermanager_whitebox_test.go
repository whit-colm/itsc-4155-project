package mockdatastore

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

// mustUsername is a helper function to create a Username from a string for testing.
// It fails the test immediately if the username is invalid.
func mustUsername(t *testing.T, s string) model.Username {
	un, err := model.UsernameFromString(s)
	if err != nil {
		t.Fatalf("failed to create username from %q: %v", s, err)
	}
	return un
}

func TestNewInMemoryUserManager(t *testing.T) {
	repo := NewInMemoryUserManager()
	assert.NotNil(t, repo)
	assert.Empty(t, repo.users)
	assert.Empty(t, repo.byGithubID)
	assert.Empty(t, repo.byUsername)
}

func TestUserRepo_Create(t *testing.T) {
	t.Run("GeneratesID", func(t *testing.T) {
		repo := NewInMemoryUserManager()
		user := &model.User{
			GithubID: "test",
			Username: mustUsername(t, "user#1234"),
		}
		err := repo.Create(context.Background(), user)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, user.ID)
	})

	t.Run("WithExistingID", func(t *testing.T) {
		repo := NewInMemoryUserManager()
		id := uuid.New()
		user1 := &model.User{
			ID:       id,
			GithubID: "user1",
			Username: mustUsername(t, "user1#1234"),
		}
		assert.NoError(t, repo.Create(context.Background(), user1))

		user2 := &model.User{
			ID:       id,
			GithubID: "user2",
			Username: mustUsername(t, "user2#5678"),
		}
		assert.NoError(t, repo.Create(context.Background(), user2))

		fetched, err := repo.GetByID(context.Background(), id)
		assert.NoError(t, err)
		assert.Equal(t, "user2", fetched.GithubID)
	})

	t.Run("IndexesPopulated", func(t *testing.T) {
		repo := NewInMemoryUserManager()
		user := &model.User{
			GithubID: "ghid123",
			Username: mustUsername(t, "user#1234"),
		}
		assert.NoError(t, repo.Create(context.Background(), user))

		fetchedGithub, err := repo.GetByGithubID(context.Background(), "ghid123")
		assert.NoError(t, err)
		assert.Equal(t, user.ID, fetchedGithub.ID)

		fetchedUsername, err := repo.GetByUsername(context.Background(), user.Username)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, fetchedUsername.ID)
	})

	t.Run("DuplicateGithubIDOverwritesIndex", func(t *testing.T) {
		repo := NewInMemoryUserManager()

		user1 := &model.User{
			GithubID: "ghid",
			Username: mustUsername(t, "user1#1234"),
		}
		repo.Create(context.Background(), user1)

		user2 := &model.User{
			GithubID: "ghid",
			Username: mustUsername(t, "user2#5678"),
		}
		repo.Create(context.Background(), user2)

		fetched, err := repo.GetByGithubID(context.Background(), "ghid")
		assert.NoError(t, err)
		assert.Equal(t, user2.ID, fetched.ID)
	})
}

func TestUserRepo_GetByID(t *testing.T) {
	t.Run("Exists", func(t *testing.T) {
		repo := NewInMemoryUserManager()
		user := &model.User{
			ID:       uuid.New(),
			GithubID: "ghid",
			Username: mustUsername(t, "user#1234"),
		}
		repo.Create(context.Background(), user)

		fetched, err := repo.GetByID(context.Background(), user.ID)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, fetched.ID)
	})

	t.Run("NotExists", func(t *testing.T) {
		repo := NewInMemoryUserManager()
		_, err := repo.GetByID(context.Background(), uuid.New())
		assert.Equal(t, repository.ErrorNotFound, err)
	})
}

func TestUserRepo_Update(t *testing.T) {
	t.Run("ExistingUser", func(t *testing.T) {
		repo := NewInMemoryUserManager()
		user := &model.User{
			GithubID:    "old_ghid",
			Username:    mustUsername(t, "olduser#1234"),
			DisplayName: "Old Name",
		}
		repo.Create(context.Background(), user)

		user.GithubID = "new_ghid"
		user.Username = mustUsername(t, "newuser#5678")
		user.DisplayName = "New Name"

		updatedUser, err := repo.Update(context.Background(), user)
		assert.NoError(t, err)
		assert.Equal(t, "New Name", updatedUser.DisplayName)

		fetched, err := repo.GetByID(context.Background(), user.ID)
		assert.NoError(t, err)
		assert.Equal(t, "New Name", fetched.DisplayName)

		fetchedByNewGhid, err := repo.GetByGithubID(context.Background(), "new_ghid")
		assert.NoError(t, err)
		assert.Equal(t, user.ID, fetchedByNewGhid.ID)

		// Check old GitHub ID still exists (current implementation behavior)
		fetchedByOldGhid, err := repo.GetByGithubID(context.Background(), "old_ghid")
		assert.NoError(t, err)
		assert.Equal(t, user.ID, fetchedByOldGhid.ID)
		assert.Equal(t, "new_ghid", fetchedByOldGhid.GithubID)
	})

	t.Run("NonExistentUser", func(t *testing.T) {
		repo := NewInMemoryUserManager()
		user := &model.User{ID: uuid.New()}
		_, err := repo.Update(context.Background(), user)
		assert.Equal(t, repository.ErrorNotFound, err)
	})
}

func TestUserRepo_Delete(t *testing.T) {
	t.Run("ExistingUser", func(t *testing.T) {
		repo := NewInMemoryUserManager()
		user := &model.User{
			GithubID: "ghid",
			Username: mustUsername(t, "user#1234"),
		}
		repo.Create(context.Background(), user)

		err := repo.Delete(context.Background(), user.ID)
		assert.NoError(t, err)

		_, err = repo.GetByID(context.Background(), user.ID)
		assert.Equal(t, repository.ErrorNotFound, err)

		_, err = repo.GetByGithubID(context.Background(), "ghid")
		assert.Equal(t, repository.ErrorNotFound, err)

		_, err = repo.GetByUsername(context.Background(), user.Username)
		assert.Equal(t, repository.ErrorNotFound, err)
	})

	t.Run("NonExistentUser", func(t *testing.T) {
		repo := NewInMemoryUserManager()
		err := repo.Delete(context.Background(), uuid.New())
		assert.Equal(t, repository.ErrorNotFound, err)
	})
}

func TestUserRepo_GetByGithubID(t *testing.T) {
	t.Run("Exists", func(t *testing.T) {
		repo := NewInMemoryUserManager()
		user := &model.User{GithubID: "ghid"}
		repo.Create(context.Background(), user)

		fetched, err := repo.GetByGithubID(context.Background(), "ghid")
		assert.NoError(t, err)
		assert.Equal(t, user.ID, fetched.ID)
	})

	t.Run("NotExists", func(t *testing.T) {
		repo := NewInMemoryUserManager()
		_, err := repo.GetByGithubID(context.Background(), "nonexistent")
		assert.Equal(t, repository.ErrorNotFound, err)
	})
}

func TestUserRepo_GetByUsername(t *testing.T) {
	t.Run("Exists", func(t *testing.T) {
		repo := NewInMemoryUserManager()
		username := mustUsername(t, "user#1234")
		user := &model.User{Username: username}
		repo.Create(context.Background(), user)

		fetched, err := repo.GetByUsername(context.Background(), username)
		assert.NoError(t, err)
		assert.Equal(t, user.ID, fetched.ID)
	})

	t.Run("NotExists", func(t *testing.T) {
		repo := NewInMemoryUserManager()
		username := mustUsername(t, "nonexistent#0001")
		_, err := repo.GetByUsername(context.Background(), username)
		assert.Equal(t, repository.ErrorNotFound, err)
	})
}

func TestUserRepo_Permissions(t *testing.T) {
	t.Run("AdminTrue", func(t *testing.T) {
		repo := NewInMemoryUserManager()
		user := &model.User{Admin: true}
		repo.Create(context.Background(), user)

		admin, err := repo.Permissions(context.Background(), user.ID)
		assert.NoError(t, err)
		assert.True(t, admin)
	})

	t.Run("AdminFalse", func(t *testing.T) {
		repo := NewInMemoryUserManager()
		user := &model.User{Admin: false}
		repo.Create(context.Background(), user)

		admin, err := repo.Permissions(context.Background(), user.ID)
		assert.NoError(t, err)
		assert.False(t, admin)
	})

	t.Run("UserNotFound", func(t *testing.T) {
		repo := NewInMemoryUserManager()
		_, err := repo.Permissions(context.Background(), uuid.New())
		assert.Equal(t, repository.ErrorNotFound, err)
	})
}

func TestUserRepo_Concurrency(t *testing.T) {
	repo := NewInMemoryUserManager()
	var wg sync.WaitGroup
	count := 100

	wg.Add(count)
	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()
			user := &model.User{
				GithubID: fmt.Sprintf("%d", i+1),
				Username: mustUsername(t, fmt.Sprintf("user%d#%04d", i, i+1)),
			}
			assert.NoError(t, repo.Create(context.Background(), user))
		}(i)
	}
	wg.Wait()

	for i := 0; i < count; i++ {
		ghid := fmt.Sprintf("%d", i+1)
		user, err := repo.GetByGithubID(context.Background(), ghid)
		assert.NoError(t, err)
		assert.Equal(t, ghid, user.GithubID)
	}
}
