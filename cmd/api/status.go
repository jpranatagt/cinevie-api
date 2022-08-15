package main

import (
  "fmt"
  "net/http"
)

// response method with status, operating environment and version
func (app *application) statusHandler(w http.ResponseWriter, r *http.Request) {
  js := `{"status": "available", "environment": %q, "version": %q}`
  js = fmt.Sprintf(js, app.config.env, version)

  // response header as json
  w.Header().Set("Content-Type", "application/json")

  // http response body
  w.Write([]byte(js))
}
