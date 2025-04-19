package model

import "github.com/google/uuid"

const AuthorApiVersion string = "author.itsc-4155-group-project.edu.whits.io/v1alpha2"

type Author struct {
	ID         uuid.UUID   `json:"id"`
	GivenName  string      `json:"givenname,omitempty"`
	FamilyName string      `json:"familyname"`
	Bio        string      `json:"bio,omitempty"`
	ExtIDs     []AuthorIDs `json:"externalIdentifiers,omitempty"`
}

func (a Author) APIVersion() string {
	return AuthorApiVersion
}

type AuthorIDs struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}
