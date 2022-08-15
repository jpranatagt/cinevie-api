package main

import (
  "encoding/json"
  "errors"
	"fmt"
	"io"
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

// read client JSON request body and return custom error if only if there's one
func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {
  // Decode the body and assigned to target destination struct
  err := json.NewDecoder(r.Body).Decode(dst)
  if err != nil {
    var syntaxError *json.SyntaxError
    var unmarshalTypeError *json.UnmarshalTypeError
    var invalidUnmarshalError *json.InvalidUnmarshalError

    switch {
      // check the error type and process it accordingly
      // using errors.As()
    case errors.As(err, &syntaxError):
      return fmt.Errorf("body contain badly-formed JSON (at character %d)", syntaxError.Offset)

      // Unexpected io error in Decode()
    case errors.Is(err, io.ErrUnexpectedEOF):
      return errors.New("body contains badly-formed JSON")

    // json.UnmarshallTypeError occur when the JSON value is the wrong type
    case errors.As(err, &unmarshalTypeError):
      if unmarshalTypeError.Field != "" {
        return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
      }

      return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

      // io.EOF will be returned by Decode() if the request body is empty
    case errors.Is(err, io.EOF):
      return errors.New("body must not empty")

      // pass a non nil pointer into Decode(), something wrong in our code
    case errors.As(err, &invalidUnmarshalError):
      panic(err)

    default:
      return err
    }
  }

  return nil
}
