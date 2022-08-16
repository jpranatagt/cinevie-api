package main

import (
  "errors"
  "net/http"

  "api.cinevie.jpranata.tech/internal/data"
  "api.cinevie.jpranata.tech/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {
  // an anonymous struct to hold the expected data from the request body
  var input struct {
    Name      string    `json:"name"`
    Email     string    `json:"email"`
    Password  string    `json:"password"`
  }

  // parse the request body to the anonymous struct
  err := app.readJSON(w, r, &input)
  if err != nil {
    app.badRequestResponse(w, r, err)

    return
  }

  // copy the data from the request body into a new User struct
  user := &data.User {
    Name: input.Name,
    Email: input.Email,
    Activated: false,
  }

  // Password.Set() to generate and store the hashed and plain text passwords
  err = user.Password.Set(input.Password)
  if err != nil {
    app.serverErrorResponse(w, r, err)

    return
  }

  v := validator.New()

  // validate the user struct and return the error messages
  if data.ValidateUser(v, user); !v.Valid() {
    app.failedValidationResponse(w, r, v.Errors)

    return
  }

  // inert the user data into database
  err = app.models.Users.Insert(user)
  if err != nil {
    switch {
    case errors.Is(err, data.ErrDuplicateEmail):
      v.AddError("email", "a user with this email address already exists.")
      app.failedValidationResponse(w, r, v.Errors)
    default:
      app.serverErrorResponse(w, r, err)
    }

    return
  }

	// a goroutine sends the welcome email in the background to reduce latency
	app.background(func() {
		err = app.mailer.Send(user.Email, "user_welcome.tmpl", user)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

  // change status code to Accepted 202 since it only being processed
  err = app.writeJSON(w, http.StatusAccepted, envelope{"user": user}, nil)
  if err != nil {
    app.serverErrorResponse(w, r, err)
  }
}
