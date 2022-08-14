package main

import (
  "fmt"
  "net/http"
  "strconv"

  "github.com/julienschmidt/httprouter"
)

// POST method with /v1/movies endpoint
func (app *application) createMovieHandler(writer http.ResponseWriter, response *http.Request) {
  fmt.Fprintln(writer, "create a new movie")
}

// GET method with /v1/movies/:id endpoint
func (app *application) showMovieHandler(writer http.ResponseWriter, response *http.Request) {
  // retrieve a slice containing URL parameter
  params := httprouter.ParamsFromContext(response.Context())

  // get id using ByName and return 404 Not Found response if id is invalid
  id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
  if err != nil || id < 1 {
    http.NotFound(writer, response)
    return
  }

  fmt.Fprintf(writer, "show the details of movie %d\n", id)
}
