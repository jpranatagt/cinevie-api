package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"api.cinevie.jpranata.tech/internal/validator"
	"github.com/lib/pq"
)

// anotate the Movie struct with struct tags
// to control how they appear in the JSON encoded output
// use snake_case for the keys instead of CamelCase
// add directive "-" to hide a field and "omitempty" if only if it's empty
type Movie struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"-"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Cover       string    `json:"cover,omitempty"`
	Trailer     string    `json:"trailer,omitempty"`
	Year        int32     `json:"year,omitempty"`
	Runtime     int32     `json:"runtime,omitempty"`
	Genres      []string  `json:"genres,omitempty"`
	Stars       []string  `json:"stars,omitempty"`
	Version     int32     `json:"version"`
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

	// context with 3 seconds timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	// passing the args and scanning the system generated id, created_at, and version
	// into movie struct, QueryRow() in use since returning a system-generated row
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&movie.ID, &movie.CreatedAt, &movie.Version)
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

	// use empty context.Background() as the parent context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	// defer to make sure cancellation the context happen before the
	// Get() method returns
	defer cancel()

	// QueryRow() for returning single row (specific to a movie)
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
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
	// add 'AND version' clause as a base for updating the record in SQL query
	// preventing data race
	query := `
    UPDATE movies
    SET title = $1, description = $2, cover = $3, trailer = $4, year = $5, runtime = $6, genres = $7, stars = $8, version = version + 1
		 WHERE id = $9 AND version = $10
    RETURNING version
  `

	// args slice containing values for the placeholder parameters
	args := []interface{}{
		movie.Title,
		movie.Description,
		movie.Cover,
		movie.Trailer,
		movie.Year,
		movie.Runtime,
		pq.Array(movie.Genres),
		pq.Array(movie.Stars),
		movie.ID,
		movie.Version, // add expected movie version
	}

	// if no matching row found then the movie version has changed
	err := m.DB.QueryRow(query, args...).Scan(&movie.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

// delete
func (m MovieModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	// query to delete the record
	query := `
    DELETE FROM movies
    WHERE id = $1
  `

	// context with 3 seconds timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	defer cancel()

	// the Exec() method returns a sql.Result object
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	// get the number of rows that being affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	// if no rows being effected, it means the movies table didn't contain
	// a record with the provided ID
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

// fetch all movies
func (m MovieModel) GetAll(title string, genres []string, filters Filters) ([]*Movie, Metadata, error) {
	// use count(*) OVER() to calculate total records according to filter which being applied
	// query to retrieve all movies
	query := fmt.Sprintf(`
		SELECT count(*) OVER(), id, created_at, title, description, cover, trailer, year, runtime, genres, stars, version
    FROM movies
		WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) or $1 = '')
    AND (genres @> $2 OR $2 = '{}')
		ORDER BY %s %s, id ASC
		LIMIT $3 OFFSET $4
  `, filters.sortColumn(), filters.sortDirection())
	// context timeout in 3 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// put args as slice
	args := []interface{}{
		title,
		pq.Array(genres),
		filters.limit(),
		filters.offset(),
	}

	// returns a sql.Rows() resultset
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	// defer a call to rows.Close() to ensure the resultset
	// is closed before GetAll() returns
	defer rows.Close()

	// total records initialize with 0
	totalRecords := 0

	// initialize an empty slice to hold the movie data
	movies := []*Movie{}

	// use rows.Next to iterate through the rows in resultset
	for rows.Next() {
		// hold individual movie
		var movie Movie

		// scan the values from the row into the Movie struct
		err := rows.Scan(
			&totalRecords,
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

		if err != nil {
			return nil, Metadata{}, err
		}

		// add the Movie struct to the slice
		movies = append(movies, &movie)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	// generate a Metadata struct passing request value from client
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return movies, metadata, nil
}
