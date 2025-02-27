package testhelper

import (
	"reflect"
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

// Bad way to test equivalence of two books
//
// Be aware this isn't very efficient, that's why it's a testhelper
// function and not a model method.
func IsBookEquals(b1, b2 model.Book) bool {
	if b1.ID != b2.ID ||
		b1.Author != b2.Author ||
		b1.Title != b2.Title ||
		b1.Published != b2.Published {
		return false
	}

	if len(b1.ISBNs) != len(b2.ISBNs) {
		return false
	}

	m1 := make(map[model.ISBN]bool)
	m2 := make(map[model.ISBN]bool)

	for _, v := range b1.ISBNs {
		m1[v] = true
	}
	for _, v := range b2.ISBNs {
		m2[v] = true
	}

	return reflect.DeepEqual(m1, m2)
}

// Bad way to test equivalence of two book slices
//
// Be aware this isn't very efficient, thus why it is a testhelper
// function and not a model method.
func IsBookSliceEquals(s1, s2 []model.Book) bool {
	if len(s1) != len(s2) {
		return false
	}

	m1 := make(map[uuid.UUID]model.Book)
	m2 := make(map[uuid.UUID]model.Book)

	for _, v := range s1 {
		m1[v.ID] = v
	}
	for _, v := range s2 {
		m2[v.ID] = v
	}

	for k, v1 := range m1 {
		v2, ok := m1[k]
		if !ok {
			return false
		}
		if !IsBookEquals(v1, v2) {
			return false
		}
	}
	return true
}
