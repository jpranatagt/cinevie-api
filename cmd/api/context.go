package main

import (
  "context"
  "net/http"

  "api.cinevie.jpranata.tech/internal/data"
)

// custom contextKey type with the underlying type string
type contextKey string

// convert the string "user" to a contextKey type and assign
// it to the userContextKey
const userContextKey = contextKey("user")

// returns a new copy of the request with the provided User
// struct added to the context
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {
  ctx := context.WithValue(r.Context(), userContextKey, user)

  return r.WithContext(ctx)
}

// use when expect there to be User struct value in the context, and if
// it doesn't exist, it will firmly be an 'unexpected' error
// as we discussed earlier in the book, it's OK to panic in those circumstances
func (app *application) contextGetUser(r *http.Request) *data.User {
  user, ok := r.Context().Value(userContextKey).(*data.User)
  if !ok {
    panic("missing user value in request context")
  }

  return user
}
