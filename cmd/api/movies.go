package main

import (
	"fmt"
  "net/http"
  "time"

	"api.cinevie.jpranata.tech/internal/data"
)

// POST method with /v1/movies endpoint
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintln(w, "create a new movie")
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
  err = app.writeJSON(w, http.StatusOK, movie, nil)
  if err != nil {
    app.logger.Println(err)
    http.Error(w, "The server encountered a problem and could not process your request.", http.StatusInternalServerError)
  }
}
