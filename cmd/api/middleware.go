package main

import (
  "fmt"
	"net"
  "net/http"
	"sync"

	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
  return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event of a panic
		// as Go unwinds the stack).
		defer func() {
		// Use the builtin recover function to check if there has been a panic or
		// not.
			if err := recover(); err != nil {
				// If there was a panic, set a "Connection: close" header on the
				// response. This acts as a trigger to make Go's HTTP server
				// automatically close the current connection after a response has been
				// sent.
				w.Header().Set("Connection", "close")
				// The value returned by recover() has the type interface{}, so we use
				// fmt.Errorf() to normalize it into an error and call our
				// serverErrorResponse() helper. In turn, this will log the error using
				// our custom Logger type at the ERROR level and send the client a 500
				// Internal Server Error response.
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
  })
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	// a mutex and a map to hold the clients' IP addresses and rate limiters
	var (
		mu sync.Mutex
		clients = make(map[string]*rate.Limiter)
	)

	// return closure which 'close over' the limiter variable
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// extract client IP address from request
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)

			return
		}

		// lock the mutex to prevent this code from being executed concurrently
		mu.Lock()

		// check if the IP address already exists in the map
		// if it doesn't, then initialize a new rate limiter
		// and add the IP address along with limiter to the map
		if _, found := clients[ip]; !found {
			clients[ip] = rate.NewLimiter(2, 4)
		}

		// call Allow() method on the rate limiter for the current IP address
		// if the request isn't allowed, unlock the mutex and send a 429 Too Many Request
		if !clients[ip].Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)

			return
		}

		// unlock mutex before calling the next handler in the chain
		mu.Unlock()

		next.ServeHTTP(w, r)
	})
}
