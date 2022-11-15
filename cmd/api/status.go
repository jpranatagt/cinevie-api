package main

import (
	"net/http"
)

// response method with status, operating environment and version
func (app *application) statusHandler(w http.ResponseWriter, r *http.Request) {
	// a map which holds the response information
	data := map[string]string{
		"status":      "available",
		"environment": app.config.env,
		"version":     version,
	}

	err := app.writeJSON(w, http.StatusOK, envelope{"status": data}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// granting permissions for browser based client to access movies resource
func (app *application) permissionsHandler(w http.ResponseWriter, r *http.Request) {
  // retrieve the user from request context
  user := app.contextGetUser(r)

  // get the slice codes of permissions for the user
  permissions, err := app.models.Permissions.GetAllForUser(user.ID)
  if err != nil {
      app.serverErrorResponse(w, r, err)

      return
  }

  // if there are no granted permission exist then response with empty string
  err = app.writeJSON(w, http.StatusOK, envelope{"permissions": permissions}, nil)
  if err != nil {
      app.serverErrorResponse(w, r, err)
  }
}

