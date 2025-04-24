package model

import (
	"io"

	"github.com/google/uuid"
)

const BlobApiVersion string = "blob.itsc-4155-group-project.edu.whits.io/v1"

type Blob struct {
	ID       uuid.UUID         `json:"id"`
	Metadata map[string]string `json:"metadata"`
	Content  io.Reader         `json:"content"`
}

func (b Blob) APIVersion() string {
	return BlobApiVersion
}
