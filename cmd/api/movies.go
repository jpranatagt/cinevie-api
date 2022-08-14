package main

import (
  "fmt"
  "net/http"
)

// POST method with /v1/movies endpoint
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintln(w, "create a new movie")
}

// GET method with /v1/movies/:id endpoint
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {
	// use readIDParam helper method
	id, err := app.readIDParam(r)
  if err != nil || id < 1 {
    http.NotFound(w, r)
    return
  }

  fmt.Fprintf(w, "show the details of movie %d\n", id)
}
