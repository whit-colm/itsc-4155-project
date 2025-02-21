package models

// album represents data about a record album.
type Book struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Author    string `json:"author"`
	Published int    `json:"publish"`
}

// books slice to seed record album data.
var Books = []Book{
	{
		"f357e3e8-b812-41ca-a93b-d3e9e5c4bd54",
		"A Tale of Two Cities",
		"Charles Dickens",
		1859,
	},
	{
		"172b68e3-02de-4d18-aa40-eef3b1d9944d",
		"The Little Prince",
		"Antoine de Saint-Exupéry",
		1943,
	},
	{
		"4a5b73d1-803d-4273-a006-8db11d6ae442",
		"The Alchemist",
		"Paulo Coelho",
		1988,
	},
}
