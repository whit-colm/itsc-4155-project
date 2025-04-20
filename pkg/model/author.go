package model

import "github.com/google/uuid"

const AuthorApiVersion string = "author.itsc-4155-group-project.edu.whits.io/v1alpha3"

type Author struct {
	ID         uuid.UUID   `json:"id"`
	GivenName  string      `json:"given_name,omitempty"`
	FamilyName string      `json:"family_name"`
	Bio        string      `json:"bio,omitempty"`
	ExtIDs     []AuthorIDs `json:"ext_ids,omitempty"`
}

func (a Author) APIVersion() string {
	return AuthorApiVersion
}

type AuthorIDs struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}
