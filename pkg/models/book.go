package models

// album represents data about a record album.
type Book struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Published int    `json:"publish"`
}

// books slice to seed record album data.
var Books = []Book{}
