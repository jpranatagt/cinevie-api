package main

import (
  "fmt"
  "net/http"
)

// logError() is a generic helper method for logging error message
func (app *application) logError(r *http.Request, err error) {
  app.logger.Println(err)
}

// generice helper method for logging JSON-formatted error
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
  env := envelope{"error": message}

  err := app.writeJSON(w, status, env, nil)
  if err != nil {
    app.logError(r, err)
    w.WriteHeader(500)
  }
}

// unexpected problem at runtime
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
  app.logError(r, err)

  message := "the server encountered a problem and could not process your request."
  app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// 404 not found
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
  message := "the requested resource could not be found."
  app.errorResponse(w, r, http.StatusNotFound, message)
}

// 405 method not allowed
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
  message := fmt.Sprintf("the %s method not supported for this resource.", r.Method)
  app.errorResponse(w, r, http.StatusMethodNotAllowed, message)
}

