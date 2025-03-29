package model

import "github.com/google/uuid"

type Blob struct {
	ID      uuid.UUID
	Content string
}
