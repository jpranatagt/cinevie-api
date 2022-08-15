package main

import (
  "net/http"
)

// response method with status, operating environment and version
func (app *application) statusHandler(w http.ResponseWriter, r *http.Request) {
  // a map which holds the response information
  data := map[string]string {
    "status": "available",
    "environment": app.config.env,
    "version": version,
  }

  err := app.writeJSON(w, http.StatusOK, data, nil)
  if err != nil {
    app.logger.Println(err)
    http.Error(w, "The server encontered a problem and could not process your request.", http.StatusInternalServerError)
  }
}
