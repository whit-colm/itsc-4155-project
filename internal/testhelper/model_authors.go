package testhelper

import (
	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
)

var ExampleAuthors []model.Author = []model.Author{
	{
		ID:         uuid.MustParse("01959161-cdfc-7142-8bab-a7008477f417"),
		GivenName:  "Charles",
		FamilyName: "Dickens",
	}, {
		ID:         uuid.MustParse("01959161-cdfc-7c45-91e3-9c785be04942"),
		GivenName:  "Antoine",
		FamilyName: "de Saint-Exupéry",
	}, {
		ID:         uuid.MustParse("01959161-cdfc-77a4-930d-0732bbf87ea6"),
		GivenName:  "Paulo",
		FamilyName: "Coelho",
	},
}

var ExampleAuthor model.Author = model.Author{
	ID:         uuid.MustParse("0124e053-3580-7000-a59a-fb9e45afdc80"),
	GivenName:  "Agatha",
	FamilyName: "Christie",
}
