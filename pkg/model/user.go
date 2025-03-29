package model

import "github.com/google/uuid"

type User struct {
	ID            uuid.UUID `json:"id"`
	GitHubID      string    `json:"github_id"`
	DisplayName   string    `json:"name"`
	UserHandle    string    `json:"username"`
	Email         string    `json:"email"`
	EmailVerified string    `json:"email_verified"`
	Avatar        uuid.UUID `json:"avatar"`
}
