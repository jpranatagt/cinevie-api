package main

import (
  "errors"
  "net/http"
	"time"

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

  // after the user record has been created in the database,
  // generate new activation token for user
  token, err := app.models.Tokens.New(user.ID, 3 * time.Hour, data.ScopeActivation)
  if err != nil {
	app.serverErrorResponse(w, r, err)

	return
  }

	// a goroutine sends the welcome email in the background to reduce latency
	app.background(func() {
		// use map to hold multiple pieces of data which would be passed to activation email
		data := map[string]interface{} {
			"activationToken": token.Plaintext,
			"userID": user.ID,
			"userName": user.Name,
		}

		err = app.mailer.Send(user.Email, "user_welcome.tmpl", data)
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

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {
	// parse the plain text activation token from the request body
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)

		return
	}

	// validate plain text token provided by client
	v := validator.New()

	if data.ValidateTokenPlaintext(v, input.TokenPlaintext); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)

		return
	}

	// retrieve the user details associated with the token
	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
				v.AddError("token", "invalid or expired activation token.")
				app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	// update the user's activation status
	user.Activated = true

	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.editConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}

		return
	}

	// if everything run successfully, then delete all activation tokens for the user
	err = app.models.Tokens.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)

		return
	}

	// send the updated user details to client in JSON response
	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
