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
