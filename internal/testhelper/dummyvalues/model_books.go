package dummyvalues

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
			model.MustNewISBN("9780451530578", model.ISBN13),
		},
		Title:     "A Tale of Two Cities",
		AuthorIDs: uuid.UUIDs{uuid.MustParse("01959161-cdfc-7142-8bab-a7008477f417")},
		Published: civil.Date{Year: 1859, Month: time.November, Day: 26},
	}, {
		ID: uuid.MustParse("0124e053-3580-7000-875a-c17e9ba5023c"),
		ISBNs: []model.ISBN{
			model.MustNewISBN("0156012197", model.ISBN10),
			model.MustNewISBN("9780156012195", model.ISBN13),
		},
		Title:     "The Little Prince",
		AuthorIDs: uuid.UUIDs{uuid.MustParse("01959161-cdfc-7c45-91e3-9c785be04942")},
		Published: civil.Date{Year: 1943, Month: time.April},
	}, {
		ID: uuid.MustParse("0124e053-3580-7000-9127-dd33bb29c893"),
		ISBNs: []model.ISBN{
			model.MustNewISBN("0062315005", model.ISBN10),
			model.MustNewISBN("9780061122415", model.ISBN13),
		},
		Title:     "The Alchemist",
		AuthorIDs: uuid.UUIDs{uuid.MustParse("01959161-cdfc-77a4-930d-0732bbf87ea6")},
		Published: civil.Date{Year: 1988},
	},
}

var ExampleBook model.Book = model.Book{
	ID: uuid.MustParse("019595bd-8d5c-75c6-b81b-07a9a7f81702"),
	ISBNs: []model.ISBN{
		model.MustNewISBN("0141439742", model.ISBN10),
		model.MustNewISBN("978-0141439747", model.ISBN13),
	},
	Title:     "Oliver Twist",
	AuthorIDs: uuid.UUIDs{uuid.MustParse("01959161-cdfc-7142-8bab-a7008477f417")},
	Published: civil.Date{Year: 1837, Month: time.February},
}

// Known dead book, will not link to any author.
var DeadBook model.Book = model.Book{
	ID: uuid.MustParse("019595bf-e65d-7ca6-a9e2-440911907a01"),
	ISBNs: []model.ISBN{
		model.MustNewISBN("0062073486", model.ISBN10),
		model.MustNewISBN("978-0062073488", model.ISBN13),
	},
	Title:     "Dream of the Red Chamber",
	AuthorIDs: uuid.UUIDs{uuid.MustParse("00000000-0000-8000-0000-100000000000")},
	Published: civil.Date{Year: 1791},
}

// Bad way to test equivalence of two books
//
// Be aware this isn't very efficient, that's why it's a testhelper
// function and not a model method.
func IsBookEquals(b1, b2 model.Book) bool {
	if b1.ID != b2.ID ||
		b1.Title != b2.Title ||
		b1.Published != b2.Published {
		return false
	}

	if len(b1.ISBNs) != len(b2.ISBNs) {
		return false
	}

	if !func() bool {
		m1 := make(map[model.ISBN]bool)
		m2 := make(map[model.ISBN]bool)

		for _, v := range b1.ISBNs {
			m1[v] = true
		}
		for _, v := range b2.ISBNs {
			m2[v] = true
		}

		return reflect.DeepEqual(m1, m2)
	}() {
		return false
	}

	return func() bool {
		m1 := make(map[uuid.UUID]bool)
		m2 := make(map[uuid.UUID]bool)

		for _, v := range b1.AuthorIDs {
			m1[v] = true
		}
		for _, v := range b2.AuthorIDs {
			m2[v] = true
		}

		return reflect.DeepEqual(m1, m2)
	}()
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
