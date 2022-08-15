package main

import (
	"fmt"
  "net/http"
  "time"

	"api.cinevie.jpranata.tech/internal/data"
	"api.cinevie.jpranata.tech/internal/validator"
)

// POST method with /v1/movies endpoint
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
	// an anonymous struct to be target of decode destination
  var input struct {
    Title     	string    `json:"title"`
    Description	string    `json:"description"`
    Cover     	string    `json:"cover"`
    Trailer     string   	`json:"trailer"`
    Year      	int32     `json:"year"`
    Runtime   	int32     `json:"runtime"`
    Genres    	[]string  `json:"genres"`
    Cast    		[]string  `json:"cast"`
  }

  // initialize json.Decoder instance to read data from request body
  // and use the Decode() method to decode the body contents into the input struct
  // must pass non-nil pointer to Decode() and it'll return error at runtime
  err := app.readJSON(w, r, &input)
  if err != nil {
    app.badRequestResponse(w, r, err)

    return
  }

	// initialize new validator instance
  v := validator.New()

	v.Check(input.Title != "", "title", "must be provided")
	v.Check(len(input.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(input.Year != 0, "year", "must be provided")
	v.Check(input.Year >= 1888, "year", "must be greater than 1888")
	v.Check(input.Year <= int32(time.Now().Year()), "year", "must not be in the future")
	v.Check(input.Runtime != 0, "runtime", "must be provided")
	v.Check(input.Runtime > 0, "runtime", "must be a positive integer")
	v.Check(input.Genres != nil, "genres", "must be provided")
	v.Check(len(input.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(input.Genres) <= 5, "genres", "must not contain more than 5 genres")
	// Note that we're using the Unique helper in the line below to check that all
	// values in the input.Genres slice are unique.
	v.Check(validator.Unique(input.Genres), "genres", "must not contain duplicate values")
	// Use the Valid() method to see if any of the checks failed. If they did, then use
	// the failedValidationResponse() helper to send a response to the client, passing
	// in the v.Errors map.
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)

		return
	}

  // no need to close r.Body since it'll done by http.Server automatically
  fmt.Fprintf(w, "%+v\n", input)
}

// GET method with /v1/movies/:id endpoint
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
  // use readIDParam helper function
  id, err := app.readIDParam(r)
  if err != nil || id < 1 {
    http.NotFound(w, r)
    return
  }

  // a new instance of the Movie struct
  // year field hasn't been specified yet
  movie := data.Movie {
    ID:         	id,
    CreatedAt:  	time.Now(),
    Title:      	"The Shawshank Redemption",
		Description:	"Two imprisoned men bond over a number of years, finding solace and eventual redemption through acts of common decency.",
		Cover:				"https://m.media-amazon.com/images/M/MV5BMDFkYTc0MGEtZmNhMC00ZDIzLWFmNTEtODM1ZmRlYWMwMWFmXkEyXkFqcGdeQXVyMTMxODk2OTU@._V1_FMjpg_UX674_.jpg",
		Trailer:			"https://www.youtube.com/watch?v=6hB3S9bIaco",
		Year:					1994,
    Runtime:    	142,
    Genres:     	[]string{"drama"},
    Cast:     		[]string{"Tim Robbins", "Morgan Freeman", "Bob Gunton"},
    Version:    	1,
  }

  // encode struct into JSON
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
  if err != nil {
    app.logger.Println(err)
    http.Error(w, "The server encountered a problem and could not process your request.", http.StatusInternalServerError)
  }
}
