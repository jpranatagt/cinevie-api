package main

import (
  "fmt"
  "net/http"
)

func (app *application) statusHandler(writer http.ResponseWriter, response *http.Request) {
  fmt.Fprintln(writer, "status: available")
  fmt.Fprintf(writer, "environment: %s\n", app.config.env)
  fmt.Fprintf(writer, "version: %s\n", version)
}
