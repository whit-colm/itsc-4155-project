package model

import (
	"io"

	"github.com/google/uuid"
)

type Blob struct {
	ID      uuid.UUID
	Content io.Reader
}
