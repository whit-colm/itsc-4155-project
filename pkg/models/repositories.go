package models

import "github.com/google/uuid"

type BookRepository interface {
	CreateBook(book *Book) error
	GetBookByID(id uuid.UUID) (*Book, error)
}
