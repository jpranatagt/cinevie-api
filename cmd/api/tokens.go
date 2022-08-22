package main

import (
//  	"fmt"
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
		v.AddError("email", "user has already been activated.")
		app.failedValidationResponse(w, r, v.Errors)

		return
	}

	// otherwise create an new activation token
	token, err := app.models.Tokens.New(user.ID, 3*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)

		return
	}

	// email user with their additional activation token
	app.background(func() {
		data := map[string]interface{}{
			"activationToken": token.Plaintext,
			"userName":        user.Name,
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
		Email    string `json:"email"`
		Password string `json:"password"`
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

	// if doesn't match return invalidCrendentialsResponse with 401 code
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

    // delete the previous token if it exists
	err = app.models.Tokens.DeleteAllForUser(data.ScopeAuthentication, user.ID)

	// generate a new token with the 24 hours expiry and the scope 'authentication'
	token, err := app.models.Tokens.New(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// encode the token to JSON and send it in the response along with a 201 status code
	/* http.SetCookie(w, &http.Cookie{
	  Name: "session_token",
	  Value: token.Plaintext,
	  Expires: token.Expiry,
	})
	fmt.Printf("token type is %s\n", token.Plaintext)
	*/

	err = app.writeJSON(w, http.StatusCreated, envelope{"authentication_token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

// generate a password reset and send it to the user's email address
func (app *application) createPasswordResetTokenHandler(w http.ResponseWriter, r *http.Request) {
	// parse and validate the user's email address
	var input struct {
		Email string `json:"email"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)

		return
	}

	v := validator.New()

	if data.ValidateEmail(v, input.Email); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)

		return
	}

	// retrieve the corresponding user record for the email address
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

	// return an error if the user is not activated
	if !user.Activated {
		v.AddError("email", "user account must be activated")
		app.failedValidationResponse(w, r, v.Errors)

		return
	}

    // delete the previous token if it exists
	err = app.models.Tokens.DeleteAllForUser(data.ScopePasswordReset, user.ID)

	// otherwise, create a new password reset token with a 45-minute expiry time
	token, err := app.models.Tokens.New(user.ID, 45*time.Minute, data.ScopePasswordReset)
	if err != nil {
		app.serverErrorResponse(w, r, err)

		return
	}

	// email the user with their password reset token
	app.background(func() {
		data := map[string]interface{}{
			"passwordResetToken": token.Plaintext,
			"userName":           user.Name,
		}

		err = app.mailer.Send(user.Email, "token_password_reset.tmpl", data)
		if err != nil {
			app.logger.PrintError(err, nil)
		}
	})

	// send a 202 Accepted response and confirmation message to the client
	env := envelope{"message": "an email would be sent to you containing password reset instructions."}

	err = app.writeJSON(w, http.StatusAccepted, env, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
