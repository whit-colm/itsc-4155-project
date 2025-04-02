package model

import (
	"time"

	"github.com/google/uuid"
)

// A summary of a full user
type CommentUser struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"name,omitempty"`
	Pronouns    string    `json:"pronouns,omitempty"`
	Username    Username  `json:"username"`
	Avatar      uuid.UUID `json:"avatar,omitempty"`
}

type Comment struct {
	ID      uuid.UUID   `json:"id"`
	Book    uuid.UUID   `json:"bookID"`
	Poster  CommentUser `json:"poster"`
	Date    time.Time   `json:"date"`
	Body    string      `json:"body"`
	Reply   uuid.UUID   `json:"reply,omitempty"`
	Votes   int         `json:"votes,omitempty"`
	Deleted bool        `json:"deleted,omitempty"`
}
