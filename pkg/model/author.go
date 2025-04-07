package model

import "github.com/google/uuid"

type Author struct {
	ID         uuid.UUID `json:"id"`
	GivenName  string    `json:"givenname"`
	FamilyName string    `json:"familyname"`
}
