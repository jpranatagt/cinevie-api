package main

import (
	"expvar" // debug metrics
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

	// application condition and info
  router.HandlerFunc(http.MethodGet, "/v1/status", app.statusHandler)

	// debug metrics
  router.Handler(http.MethodGet, "/metrics", expvar.Handler())

	// movies
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermission("movies:read", app.listMoviesHandler))
  router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission("movies:write", app.createMovieHandler))
  router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermission("movies:read", app.showMovieHandler))
  router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission("movies:write", app.updateMovieHandler))
  router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission("movies:write", app.deleteMovieHandler))

	//users
  router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
  router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	// tokens
	router.HandlerFunc(http.MethodPost, "/v1/tokens/activation", app.createActivationTokenHandler)
router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// add the enableCORS() middleware
  // put it before rateLimit to prevent request exceeded of 429 too many request response
	return app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router)))))
}
