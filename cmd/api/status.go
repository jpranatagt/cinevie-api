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

// granting permission for browser based client to access movies resource
func (app *application) permissionHandler(w http.ResponseWriter, r *http.Request) {
  // retrieve the user from request context
  user := app.contextGetUser(r)

  // get the slice codes of permissions for the user
  permissions, err := app.models.Permissions.GetAllForUser(user.ID)
  if err != nil {
      app.serverErrorResponse(w, r, err)

      return
  }

  var grantedPermission = ""

  moviesRead := "movies:read"
  moviesWrite := "movies:write"

  // check if the slice includes the specified permission
  if permissions.Include(moviesRead) {
    grantedPermission = moviesRead
  }

  if permissions.Include(moviesWrite) {
    grantedPermission = moviesWrite
  }

  // if there are no granted permission exist then response with empty string
  err = app.writeJSON(w, http.StatusOK, envelope{"permission": grantedPermission}, nil)
  if err != nil {
      app.serverErrorResponse(w, r, err)
  }
}

