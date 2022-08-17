package main

import (
  "errors"
	"net/http"
  "time"

  "api.cinevie.jpranata.tech/internal/data"
  "api.cinevie.jpranata.tech/internal/validator"
)

func (app *application) createActivationTokenHandler(w http.ResponseWriter, r *http.Request) {
  // parse and validate user's email address
  var input struct {
    Email string `json:"email"`
  }

  err := app.readJSON(w, r, &input)
  if err != nil {
    app.badRequestResponse(w, r, err)
  }

  v := validator.New()

  if data.ValidateEmail(v, input.Email); !v.Valid() {
    app.failedValidationResponse(w, r, v.Errors)

    return
  }

  // retrieve corresponding record for the email address
  user, err := app.models.Users.GetByEmail(input.Email)
  if err != nil {
    switch {
    case errors.Is(err, data.ErrRecordNotFound):
      v.AddError("email", "no matching email address found.")
      app.failedValidationResponse(w, r, v.Errors)
    default:
      app.serverErrorResponse(w, r, err)
    }

    return
  }

  // return an error if the user has already been activated
  if user.Activated {
    v.AddError("email", "user has already been deactivated.")
    app.failedValidationResponse(w, r, v.Errors)

    return
  }

  // otherwise create an new activation token
  token, err := app.models.Tokens.New(user.ID, 3 * time.Hour, data.ScopeActivation)
  if err != nil {
    app.serverErrorResponse(w, r, err)

    return
  }

  // email user with their additional activation token
  app.background(func() {
    data := map[string]interface{} {
      "activationToken": token.Plaintext,
		"userName": user.Name,
    }

    // send lowercase email which has been stored in the database
    err = app.mailer.Send(user.Email, "token_activation.tmpl", data)
    if err != nil {
      app.logger.PrintError(err, nil)
    }
  })

  // send 202 Accepted response and configuration message to the client
  env := envelope{"message": "an email will be sent to you containing activation instructions."}

  err = app.writeJSON(w, http.StatusAccepted, env, nil)
  if err != nil {
    app.serverErrorResponse(w, r, err)
  }
}

func (app *application) createAuthenticationTokenHandler(w http.ResponseWriter, r *http.Request) {
  // parse the email and password from the request body
  var input struct {
    Email     string  `json:"email"`
    Password  string  `json:"password"`
  }

  err := app.readJSON(w, r, &input) // parse and assign them to input struct
  if err != nil {
    app.badRequestResponse(w, r, err)

    return
  }

  // validate the email and password
  v := validator.New()
  data.ValidateEmail(v, input.Email)
  data.ValidatePasswordPlaintext(v, input.Password)

  if !v.Valid() {
    app.failedValidationResponse(w, r, v.Errors)

    return
  }

  // if not matching return invalidCrendentialsResponse with 401 code
  user, err := app.models.Users.GetByEmail(input.Email)
  if err != nil {
    switch {
      case errors.Is(err, data.ErrRecordNotFound):
        app.invalidCredentialsResponse(w, r)
      default:
        app.serverErrorResponse(w, r, err)
    }

  return
  }

  // check if password matches
  match, err := user.Password.Matches(input.Password)
  if err != nil {
    app.serverErrorResponse(w, r, err)

    return
  }

  // if the password didn't match then call invalidCredentialsResponse again
  if !match {
    app.invalidCredentialsResponse(w, r)

    return
  }

  // generate a new token with the 24 hours expiry and the scope 'authentication'
  token, err := app.models.Tokens.New(user.ID, 3 * time.Hour, data.ScopeAuthentication)
  if err != nil {
    app.serverErrorResponse(w, r, err)
    return
  }

  // encode the token to JSON and send it in the response along with a 201 status code
  err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
  if err != nil {
    app.serverErrorResponse(w, r, err)
  }
}
