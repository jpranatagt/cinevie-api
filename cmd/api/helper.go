package main

import (
  "encoding/json"
  "errors"
  "net/http"
  "strconv"

  "github.com/julienschmidt/httprouter"
)

func (app *application) readIDParam(r *http.Request) (int64, error) {
  // parsing the parameter
  params := httprouter.ParamsFromContext(r.Context())

  id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
  if err != nil || id < 1 {
    return 0, errors.New("invalid id parameter")
  }

  return id, nil
}

// wrap the encoded JSON with parent key name of data
// it's a self documenting, clarity about what data is
// about and mitigate a security vulnerability in older browser
type envelope map[string]interface{}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
  // encode the data into JSON return error if there was one
  js, err := json.Marshal(data)
  if err != nil {
    return err
  }

  // append new line for terminal view
  js = append(js, '\n')

  // loop header map and add each header to http.ResponseWriter header
  for key, value := range headers {
    w.Header()[key] = value
  }

  // Add the content type header, write status code and JSON response
  w.Header().Set("Content-Type", "application/json")
  w.WriteHeader(status)
  w.Write(js)

  return nil
}
