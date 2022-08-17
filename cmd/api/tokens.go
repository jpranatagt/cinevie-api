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
  token, err := app.models.Tokens.New(user.ID, 24 * time.Hour, data.ScopeActivation)
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
