package mockdatastore

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

// BlobRepo implements BlobManager.
type BlobRepo struct {
	mut   sync.RWMutex
	blobs map[uuid.UUID]*model.Blob
}

var _ repository.BlobManager = (*BlobRepo)(nil)

func NewInMemoryBlobManager() *BlobRepo {
	return &BlobRepo{
		blobs: make(map[uuid.UUID]*model.Blob),
	}
}

func (m *BlobRepo) Create(ctx context.Context, blob *model.Blob) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	data, err := io.ReadAll(blob.Content)
	if err != nil {
		return fmt.Errorf("read blob content: %w", err)
	}

	// Store content as a new bytes.Reader for each retrieval
	m.blobs[blob.ID] = &model.Blob{
		ID:       blob.ID,
		Metadata: blob.Metadata,
		Content:  bytes.NewReader(data),
	}
	return nil
}

func (m *BlobRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Blob, error) {
	m.mut.RLock()
	defer m.mut.RUnlock()

	blob, exists := m.blobs[id]
	if !exists {
		return nil, fmt.Errorf("blob not found")
	}

	// Return a copy with a fresh reader
	data, _ := io.ReadAll(blob.Content)
	return &model.Blob{
		ID:       blob.ID,
		Metadata: blob.Metadata,
		Content:  bytes.NewReader(data),
	}, nil
}

// Delete implements repository.BlobManager.
func (m *BlobRepo) Delete(ctx context.Context, blobID uuid.UUID) error {
	m.mut.Lock()
	defer m.mut.Unlock()

	delete(m.blobs, blobID)
	return nil
}

// Update implements repository.BlobManager.
func (m *BlobRepo) Update(ctx context.Context, to *model.Blob) (*model.Blob, error) {
	m.mut.Lock()
	defer m.mut.Unlock()

	m.blobs[to.ID] = to
	return to, nil
}
