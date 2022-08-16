package data

import (
	"database/sql"
  "time"

	"api.cinevie.jpranata.tech/internal/validator"
)

// anotate the Movie struct with struct tags
// to control how they appear in the JSON encoded output
// use snake_case for the keys instead of CamelCase
// add directive "-" to hide a field and "omitempty" if only if it's empty
type Movie struct {
  ID          int64       `json:"id"`
  CreatedAt   time.Time   `json:"-"`
  Title       string      `json:"title"`
  Description string      `json:"description,omitempty"`
  Cover       string      `json:"cover,omitempty"`
  Trailer     string      `json:"trailer,omitempty"`
  Year        int32       `json:"year,omitempty"`
  Runtime     int32       `json:"runtime,omitempty"`
  Genres      []string    `json:"genres,omitempty"`
  Cast        []string    `json:"cast,omitempty"`
  Version     int32       `json:"version"`
}

func ValidateMovie(v *validator.Validator, movie *Movie) {
	v.Check(movie.Title != "", "title", "must be provided")
	v.Check(len(movie.Title) <= 500, "title", "must not be more than 500 bytes long")

	v.Check(movie.Description != "", "description", "must be provided")
	v.Check(len(movie.Description) <= 1500, "description", "must not be more than 1500 bytes long")

	v.Check(movie.Cover != "", "cover", "must be provided")
	v.Check(len(movie.Cover) <= 1000, "cover", "must not be more than 1000 bytes long")

	v.Check(movie.Trailer != "", "trailer", "must be provided")
	v.Check(len(movie.Trailer) <= 1000, "trailer", "must not be more than 1000 bytes long")

	v.Check(movie.Year != 0, "year", "must be provided")
	v.Check(movie.Year >= 1888, "year", "must be greater than 1888")
	v.Check(movie.Year <= int32(time.Now().Year()), "year", "must not be in the future")

	v.Check(movie.Runtime != 0, "runtime", "must be provided")
	v.Check(movie.Runtime > 0, "runtime", "must be a positive integer")

	v.Check(movie.Genres != nil, "genres", "must be provided")
	v.Check(len(movie.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(movie.Genres) <= 5, "genres", "must not contain more than 5 genres")
	// Note that we're using the Unique helper in the line below to check that all
	// values in the movie.Genres slice are unique.
	v.Check(validator.Unique(movie.Genres), "genres", "must not contain duplicate values")

	v.Check(movie.Cast != nil, "cast", "must be provided")
	v.Check(len(movie.Cast) >= 1, "cast", "must contain at least 1 genre")
	v.Check(len(movie.Cast) <= 10, "cast", "must not contain more than 10 cast")
	// Note that we're using the Unique helper in the line below to check that all
	// values in the movie.Genres slice are unique.
	v.Check(validator.Unique(movie.Cast), "cast", "must not contain duplicate values")
}

type MovieModel struct {
  DB *sql.DB
}

// creating placeholder method for CRUD process

// insert
func (m MovieModel) Insert(movie *Movie) error {
  return nil
}

// fetch
func (m MovieModel) Get(id int64) (*Movie, error) {
  return nil, nil
}

// update
func (m MovieModel) Update(movie *Movie) error {
  return nil
}

// delete
func (m MovieModel) Delete(id int64) error {
  return nil
}
