package model

import "github.com/google/uuid"

const AuthorApiVersion string = "author.itsc-4155-group-project.edu.whits.io/v1"

type Author struct {
	ID         uuid.UUID `json:"id"`
	GivenName  string    `json:"givenname"`
	FamilyName string    `json:"familyname"`
}

func (a Author) APIVersion() string {
	return AuthorApiVersion
}
