package testhelper

import (
	"context"

	"github.com/whit-colm/itsc-4155-project/pkg/repository"
)

type TestingStoreManager struct{}

var _ repository.StoreManager = (*TestingStoreManager)(nil)

// Connect implements repository.StoreManager.
func (t *TestingStoreManager) Connect(args any) error {
	return nil
}

// Disconnect implements repository.StoreManager.
func (t *TestingStoreManager) Disconnect() error {
	return nil
}

// Ping implements repository.StoreManager.
func (t *TestingStoreManager) Ping(ctx context.Context) error {
	return nil
}
