package main

import (
  "net/http"

  "github.com/julienschmidt/httprouter"
)

// update the routes() method to return a http.Handler instead of
// a *httprouter.Router
func (app *application) routes() http.Handler {
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
  router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.updateMovieHandler)
  router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.listMoviesHandler)

  router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	// use PUT method since it's idempotent after activation succeed
  // the subsequent requests would be resulted in error
  // or the state of application is not change after first succeed activation
  router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/activation", app.createActivationTokenHandler)

	return app.recoverPanic(app.rateLimit(router))
}
