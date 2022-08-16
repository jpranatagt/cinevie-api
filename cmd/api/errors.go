package main

import (
  "fmt"
  "net/http"
)

// logError() is a generic helper method for logging error message
func (app *application) logError(r *http.Request, err error) {
	// include the current request method and URL as properties in log entry
  app.logger.PrintError(err, map[string]string{
    "request_method": r.Method,
    "request_url": r.URL.String(),
  })
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

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
  app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}

// invalid validation error
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
  app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}

// race condition create conflict
func (app *application) editConflictResponse(w http.ResponseWriter, r *http.Request) {
  message := "unable to update the record due to and edit conflict, please try again"
  app.errorResponse(w, r, http.StatusConflict, message)
}
