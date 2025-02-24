package models

import (
	"time"

	"cloud.google.com/go/civil"
	"github.com/google/uuid"
)

var BooksExample []Book

func init() {
	u, err := uuid.NewV7()
	if err != nil {
		panic(-1)
	}
	BooksExample = append(BooksExample, Book{
		u,
		[]ISBNFormatter{&ISBN10{"0141439602"}, &ISBN13{"9780141439600"}},
		"A Tale of Two Cities",
		"Charles Dickens",
		civil.Date{Year: 1859, Month: time.November, Day: 26},
	})

	u, err = uuid.NewV7()
	if err != nil {
		panic(-1)
	}
	BooksExample = append(BooksExample, Book{
		u,
		[]ISBNFormatter{&ISBN10{"0156012197"}, &ISBN13{"9780156012195"}},
		"The Little Prince",
		"Antoine de Saint-Exupéry",
		civil.Date{Year: 1943, Month: time.April},
	})

	u, err = uuid.NewV7()
	if err != nil {
		panic(-1)
	}
	BooksExample = append(BooksExample, Book{
		u,
		[]ISBNFormatter{&ISBN10{"0062315005"}, &ISBN13{"9780061122415"}},
		"The Alchemist",
		"Paulo Coelho",
		civil.Date{Year: 1988},
	})
}
