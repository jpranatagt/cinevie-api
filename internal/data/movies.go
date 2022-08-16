package data

import (
	"database/sql"
	"errors"
  "time"

	"api.cinevie.jpranata.tech/internal/validator"
	"github.com/lib/pq"
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
  Stars       []string    `json:"stars,omitempty"`
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

	v.Check(movie.Stars != nil, "stars", "must be provided")
	v.Check(len(movie.Stars) >= 1, "stars", "must contain at least 1 genre")
	v.Check(len(movie.Stars) <= 10, "stars", "must not contain more than 10 stars")
	// Note that we're using the Unique helper in the line below to check that all
	// values in the movie.Genres slice are unique.
	v.Check(validator.Unique(movie.Stars), "stars", "must not contain duplicate values")
}

type MovieModel struct {
  DB *sql.DB
}

// creating placeholder method for CRUD process

// insert
func (m MovieModel) Insert(movie *Movie) error {
	// sql for inserting movie record and returning
  // the system generated data to placeholder parameters
  query := `
    INSERT INTO movies (title, description, cover, trailer, year, runtime, genres, stars)
    VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    RETURNING id, created_at, version
  `

  // args slice containing the values for the placeholder parameters
  // from movie struct and make it clear what values being used/where
  args := []interface{}{
		movie.Title,
		movie.Description,
		movie.Cover,
		movie.Trailer,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		pq.Array(movie.Stars),
	}

  // passing the args and scanning the system generated id, created_at, and version
  // into movie struct, QueryRow() in use since returning a system-generated row
  return m.DB.QueryRow(query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
}

// fetch
func (m MovieModel) Get(id int64) (*Movie, error) {
	// movie ID using bigserial type and auto incrementing at 1 by default (2, 3, 4 and so on)
  // there would be no movie ID less than 1 thus return error if that happen
  if id < 1 {
    return nil, ErrRecordNotFound // errors shortcut
  }

  // query for retrieving data
  query := `
    SELECT id, created_at, title, description, cover, trailer, year, runtime, genres, stars, version
    FROM movies
    WHERE id = $1
  `

  // a Movie struct to hold the data returned by the query
  var movie Movie

  // QueryRow() for returning single row (specific to a movie)
  err := m.DB.QueryRow(query, id).Scan(
    &movie.ID,
    &movie.CreatedAt,
    &movie.Title,
		&movie.Description,
		&movie.Cover,
		&movie.Trailer,
    &movie.Year,
    &movie.Runtime,
    pq.Array(&movie.Genres),
		pq.Array(&movie.Stars),
    &movie.Version,
  )

  // return a sql.ErrNoRows error if no matching movie found
  // use custom ErrRecordNotFound instead
  if err != nil {
    switch {
    case errors.Is(err, sql.ErrNoRows):
      return nil, ErrRecordNotFound
    default:
      return nil, err
    }
  }

  // otherwise return  a pointer to the Movie struct
  return &movie, nil
}

// update
func (m MovieModel) Update(movie *Movie) error {
  return nil
}

// delete
func (m MovieModel) Delete(id int64) error {
  return nil
}
