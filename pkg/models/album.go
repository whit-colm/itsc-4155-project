package models

// album represents data about a record album.
type Album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Rating float64 `json:"price"`
}

// albums slice to seed record album data.
var Albums = []Album{
	{ID: "1", Title: "Al Mundo Azul", Artist: "Mr Twin Sister", Rating: 4.3},
	{ID: "2", Title: "Humanz", Artist: "Gorillaz", Rating: 3.1},
	{ID: "3", Title: "Geogaddi", Artist: "Boards of Canada", Rating: 5.0},
}
