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
		app.serverErrorResponse(w, r, err)
  }
}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// extract the movie ID from URL
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)

		return
	}

	// fetch the existing movie record using the id from request
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

	// input struct to hold expected data from client
	var input struct {
		Title 			*string 				`json:"title"`
		Description *string				`json:"description"`
		Cover				*string				`json:"cover"`
		Trailer			*string				`json:"trailer"`
		Year				*int32				`json:"year"`
		Runtime			*int32				`json:"runtime"`
		Genres			[]string			`json:"genres"`
		Stars				[]string			`json:"stars"`
	}

	// read the JSON request body data into the input struct
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// if the input.Field is nil then no corresponding value provided
	// in JSON request body. Otherwise update the corresponding fields.
	if input.Title != nil {
		movie.Title = *input.Title
	}

	if input.Description != nil {
		movie.Description = *input.Description
	}

	if input.Cover != nil {
		movie.Cover = *input.Cover
	}

	if input.Trailer != nil {
		movie.Trailer = *input.Trailer
	}

	if input.Year != nil {
		movie.Year = *input.Year
	}

	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}

	if input.Genres != nil { // slice would return nil if it's empty
		movie.Genres = input.Genres // don't need to dereference a slice
	}

	if input.Stars != nil {
		movie.Stars = input.Stars
	}

	// validate the updated record
	v := validator.New()

	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)

		return
	}

	// pass the updated movie record to our new Update() method
	err = app.models.Movies.Update(movie)
	if err != nil {
	switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	// write the updated movie record in a JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)

	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// extract ID from request parameter
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)

		return
	}

	// delete and sending 404 not found if there isn't matching record
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	// return 200 OK status code along with success message
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie sucessfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// GET method with /v1/movies endpoint to show listed movies
func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	// input struct to hold expected values from request query string
	var input struct {
		Title string
		Genres []string
		data.Filters
	}

	// initialize validator instance
	v := validator.New()

	// get the url.Values map containing the query string data
	qs := r.URL.Query()

	// extract the title and genres using helper
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})

	// get the page and page_size as integers
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	// extract sort format
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// supported sort values for this endpoint to the sort safe list
	input.Filters.SortSafelist = []string{"id", "title", "year", "runtime", "-id", "-title", "-year", "-runtime"}

	// validation
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)

		return
	}

	// call GetAll() method to retrieve the movies and passing various filter parameters
	movies, metadata, err := app.models.Movies.GetAll(input.Title, input.Genres, input.Filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)

		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"metadata": metadata, "movies": movies}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

