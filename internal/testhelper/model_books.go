package testhelper

import (
	"time"

	"cloud.google.com/go/civil"
	"github.com/google/uuid"
	"github.com/whit-colm/itsc-4155-project/pkg/model"
)

var ExampleBooks []model.Book = []model.Book{
	{
		ID: uuid.MustParse("0124e053-3580-7000-8794-db4a97089840"),
		ISBNs: []model.ISBN{
			model.MustNewISBN("0141439602", model.ISBN10),
			model.MustNewISBN("9780141439600", model.ISBN13),
		},
		Title:     "A Tale of Two Cities",
		Author:    "Charles Dickens",
		Published: civil.Date{Year: 1859, Month: time.November, Day: 26},
	}, {
		ID: uuid.MustParse("0124e053-3580-7000-875a-c17e9ba5023c"),
		ISBNs: []model.ISBN{
			model.MustNewISBN("0156012197", model.ISBN10),
			model.MustNewISBN("9780156012195", model.ISBN13),
		},
		Title:     "The Little Prince",
		Author:    "Antoine de Saint-Exupéry",
		Published: civil.Date{Year: 1943, Month: time.April},
	}, {
		ID: uuid.MustParse("0124e053-3580-7000-9127-dd33bb29c893"),
		ISBNs: []model.ISBN{
			model.MustNewISBN("0062315005", model.ISBN10),
			model.MustNewISBN("9780061122415", model.ISBN13),
		},
		Title:     "The Alchemist",
		Author:    "Paulo Coelho",
		Published: civil.Date{Year: 1988},
	},
}

var ExampleBook model.Book = model.Book{
	ID: uuid.MustParse("0124e053-3580-7000-a59a-fb9e45afdc80"),
	ISBNs: []model.ISBN{
		model.MustNewISBN("0062073486", model.ISBN10),
		model.MustNewISBN("978-0062073488", model.ISBN13),
	},
	Title:     "And Then There Were None",
	Author:    "Agatha Christie",
	Published: civil.Date{Year: 1939, Month: time.November, Day: 6},
}
