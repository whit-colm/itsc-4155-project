package model

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"

	"github.com/google/uuid"
)

type Blob struct {
	ID       uuid.UUID         `json:"id"`
	Metadata map[string]string `json:"metadata"`
	Content  io.Reader         `json:"content"`
}

func (o Blob) MarshalJSON() ([]byte, error) {
	var cE []byte
	if c, err := io.ReadAll(o.Content); err != nil {
		return nil, err
	} else {
		base64.StdEncoding.Encode(c, cE)
	}
	return json.Marshal(struct {
		ID       uuid.UUID         `json:"id"`
		Metadata map[string]string `json:"metadata"`
		Content  []byte            `json:"content"`
	}{
		ID:       o.ID,
		Metadata: o.Metadata,
		Content:  cE,
	})
}

func (o *Blob) UnmarshalJSON(b []byte) error {
	var aux struct {
		ID       uuid.UUID         `json:"id"`
		Metadata map[string]string `json:"metadata"`
		Content  []byte            `json:"content"`
	}
	if err := json.Unmarshal(b, &aux); err != nil {
		return err
	}
	var c []byte
	if _, err := base64.StdEncoding.Decode(c, aux.Content); err != nil {
		return err
	}
	o.ID = aux.ID
	o.Metadata = aux.Metadata
	o.Content = bytes.NewReader(c)
	return nil
}
