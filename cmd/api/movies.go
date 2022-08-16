package main

import (
	"errors"
	"fmt"
  "net/http"

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
    Stars    		[]string  `json:"stars"`
  }

  // initialize json.Decoder instance to read data from request body
  // and use the Decode() method to decode the body contents into the input struct
  // must pass non-nil pointer to Decode() and it'll return error at runtime
  err := app.readJSON(w, r, &input)
  if err != nil {
    app.badRequestResponse(w, r, err)

    return
  }

	// copy the values from the input (put in by readJSON through pointer) struct to a new Movie struct
	// note that the movie variable contains a pointer to a Movie struct
	movie := &data.Movie {
		Title: 				input.Title,
		Description:	input.Description,
		Cover:				input.Cover,
		Trailer:			input.Trailer,
		Year: 				input.Year,
		Runtime: 			input.Runtime,
		Genres: 			input.Genres,
		Stars:				input.Stars,
	}

	// initialize a new validator
	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)

		return
	}

	// passing in a movie pointer to the validated movie struct by ValidateMovie
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)

		return
	}

	// include Location header to let the client know which URL they can find
	// newly created resource at. Make an empty http.Header map and use Set()
	// method to add new Location header
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movies/%d", movie.ID))

  // no need to close r.Body since it'll done by http.Server automatically
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
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
  /* movie := data.Movie {
    ID:         	id,
    CreatedAt:  	time.Now(),
    Title:      	"The Shawshank Redemption",
		Description:	"Two imprisoned men bond over a number of years, finding solace and eventual redemption through acts of common decency.",
		Cover:				"https://m.media-amazon.com/images/M/MV5BMDFkYTc0MGEtZmNhMC00ZDIzLWFmNTEtODM1ZmRlYWMwMWFmXkEyXkFqcGdeQXVyMTMxODk2OTU@._V1_FMjpg_UX674_.jpg",
		Trailer:			"https://www.youtube.com/watch?v=6hB3S9bIaco",
		Year:					1994,
    Runtime:    	142,
    Genres:     	[]string{"drama"},
    Stars:     		[]string{"Tim Robbins", "Morgan Freeman", "Bob Gunton"},
    Version:    	1,
  } */

	// fetch specific movie data and return custom error if it happen
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.notFoundResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
		}

		return
	}

  // encode struct into JSON
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
  if err != nil {
    app.logger.Println(err)
    http.Error(w, "The server encountered a problem and could not process your request.", http.StatusInternalServerError)
  }
}
