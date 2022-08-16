package main

import (
  "encoding/json"
  "errors"
	"fmt"
	"io"
	"net/url"
  "net/http"
  "strconv"
	"strings"

	"api.cinevie.jpranata.tech/internal/validator"

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
	 // use http.MaxBytesReader() to limit the size of the request body to 1MB
  maxBytes := 1_048_576
  r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

  // initialize the json.Decoder and call the DissallowUnknownFields() method on it before decoding it
  // the decoder will return error if body include any unmatched field with target destination (struct)
  dec := json.NewDecoder(r.Body)
  dec.DisallowUnknownFields()

  // Decode the body and assigned to target destination struct
	err := dec.Decode(dst)
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
		// extract field name from Decode() error message ("json: unknown field "<name>".")
    case strings.HasPrefix(err.Error(), "json: unknown field "):
      fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")

      return fmt.Errorf("body contains unknown key %s", fieldName)

      // request body exceeds 1MB
    case err.Error() == "http: request body too large":
      return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

      // pass a non nil pointer into Decode(), something wrong in our code
    case errors.As(err, &invalidUnmarshalError):
      panic(err)

    default:
      return err
    }
  }

	// if body only contain single JSON value
  // this will return an io.EOF error
  // if not there must be additional data being inserted
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
  }

  return nil
}


// return the default value if no matching key could be found
func (app *application) readString(qs url.Values, key string, defaultValue string) string {
  // extract key if no key exists return empty string ""
  s := qs.Get(key)

  if s == "" {
    return defaultValue
  }

  return s
}

// split string into a slice on the comma character
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
  csv := qs.Get(key)

  if csv == "" {
    return defaultValue
  }

  return strings.Split(csv, ",")
}

// reads and converts string value into integer
// use validation to return appropriate error
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
  s := qs.Get(key)

  if s == "" {
    return defaultValue
  }

  // conversion
  i, err := strconv.Atoi(s)
  if err != nil {
    v.AddError(key, "must be an integer value")

    return defaultValue
  }

  return i
}

// accept an arbitrary function with signature func()
func (app *application) background(fn func()) {
	// increment the WaitGroup process number by one
	app.wg.Add(1)
  // launch go routine
  go func() {
    // recover any panic by logging the message instead of terminating application
    defer func() {
			// use defer to decrement the WaitGroup counter before the goroutine returns
			defer app.wg.Done()

      if err := recover(); err != nil {
        app.logger.PrintError(fmt.Errorf("%s", err), nil)
      }
    }()

    // execute the arbitrary function
    fn()
  }()
}
