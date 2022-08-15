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

	// copy the values from the input struct to a new Movie struct
	movie := &data.Movie {
		Title: 				input.Title,
		Description:	input.Description,
		Cover:				input.Cover,
		Trailer:			input.Trailer,
		Year: 				input.Year,
		Runtime: 			input.Runtime,
		Genres: 			input.Genres,
		Cast:					input.Cast,
	}

	// initialize a new validator
	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
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
