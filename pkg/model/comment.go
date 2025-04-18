package model

import (
	"time"

	"github.com/google/uuid"
)

const CommentApiVersion string = "comment.itsc-4155-group-project.edu.whits.io/v1alpha1"

// A summary of a full user object containing only the parts relevant
// for a comment.
type CommentUser struct {
	ID          uuid.UUID `json:"id"`
	DisplayName string    `json:"name,omitempty"`
	Pronouns    string    `json:"pronouns,omitempty"`
	Username    Username  `json:"username"`
	Avatar      uuid.UUID `json:"avatar,omitempty"`
}

type Comment struct {
	ID   uuid.UUID `json:"id"`
	Body string    `json:"body"`
	// The ID of the book object this review is under.
	Book   uuid.UUID   `json:"bookID"`
	Date   time.Time   `json:"date"`
	Poster CommentUser `json:"poster"`

	// Either rating xor reply must be populated, but never both.
	// If rating is populated that means it's an original ("top level")
	// comment, which users are only allowed one of per-book. If reply
	// is populated, that means it's a reply to another review, which
	// as many as desired are allowed.
	Rating float32   `json:"rating,omitempty"`
	Parent uuid.UUID `json:"parent,omitempty"`

	// If a comment has been deleted, the body and author information
	// should be null, however the comment entry itself in the
	// datastore should still 'exist' as replies will still need to
	// reference it
	Deleted bool      `json:"deleted,omitempty"`
	Edited  time.Time `json:"edited,omitempty"`
	Votes   int       `json:"votes,omitempty"`
}

func (c Comment) APIVersion() string {
	return CommentApiVersion
}
