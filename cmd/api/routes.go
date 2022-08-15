package main

import (
  "net/http"

  "github.com/julienschmidt/httprouter"
)

func (app *application) routes() *httprouter.Router {
  // a new httprouter instance
  router := httprouter.New()

  // the entire error response triggered by router module
  // convert the notFoundResponse() helper into a http.Handler
  // using http.HanlderFunc adapter and set it as custom error
  router.NotFound = http.HandlerFunc(app.notFoundResponse)

  // also custom error
  router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

  // register relevant methods, URL patterns and handler functions for the endpoints
  // using HandlerFunc method
  router.HandlerFunc(http.MethodGet, "/v1/status", app.statusHandler)
  router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
  router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

  // return the httprouter instance
  return router
}
