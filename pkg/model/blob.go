package model

import (
	"io"

	"github.com/google/uuid"
)

type Blob struct {
	ID       uuid.UUID         `json:"id"`
	Metadata map[string]string `json:"metadata"`
	Content  io.Reader         `json:"content"`
}
