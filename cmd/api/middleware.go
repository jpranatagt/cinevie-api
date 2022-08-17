package main

import (
	"errors"
	"expvar"
  "fmt"
  "net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"api.cinevie.jpranata.tech/internal/data"
	"api.cinevie.jpranata.tech/internal/validator"

	"github.com/tomasen/realip"
	"github.com/felixge/httpsnoop"
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
	// struct to hold the rate limiter and last seen time for each client
	type client struct {
		limiter *rate.Limiter
		lastSeen time.Time
	}

	// a mutex and a map to hold the clients' IP addresses and rate limiters
	var (
		mu sync.Mutex
		// update the map so the values are pointers to a client struct
		clients = make(map[string]*client)
	)

	// launch goroutine in the background which removes old entries from the clients
	// once every minute
	go func() {
		for {
			time.Sleep(time.Minute)

			// lock the mutex to prevent any rate limiter checks from happening
			// while the cleanup is taking place
			mu.Lock()

			// loop through the entire clients, if they haven't been seen within
			// the last three minutes, delete the corresponding entry from the map
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}

			// unlock when the cleanup check is complete
			mu.Unlock()
		}
	}()


	// return closure which 'close over' the limiter variable
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// only process rate limiter if it is enabled
		if app.config.limiter.enabled {
		// use the realip.FromRequest() function to get the client's real IP address
			ip := realip.FromRequest(r)

		// lock the mutex to prevent this code from being executed concurrently
		mu.Lock()

		if _, found := clients[ip]; !found {
				// use the request per second and burst value from config struct
				clients[ip] = &client{limiter: rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)}
		}

		// update the last seen time for the client
		clients[ip].lastSeen = time.Now()

		// call Allow() method on the rate limiter for the current IP address
		// if the request isn't allowed, unlock the mutex and send a 429 Too Many Request
		if !clients[ip].limiter.Allow() {
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)

			return
		}

		// unlock mutex before calling the next handler in the chain
		mu.Unlock()
	}

		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		// indicates to any caches that the response may "vary" based on the value
		// of the Authorization header in the request
		w.Header().Add("Vary", "Authorization")

		// retrieve the value of the Authorization header from request
		// return the empty string "" if there is no such header found
		authorizationHeader := r.Header.Get("Authorization")

		// if there is no Authorization header found, use the contextSetUser() helper
		// to add the AnonymousUser to the request context, then call the next handler
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)

			return
		}

		// otherwise, expect the Authorization header to be in the format
		// "Bearer <token>", split this into its constituent parts
		// and if the header isn't in the expected format return 401 Unauthorized
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)

			return
		}

		// extract the actual authentication token from the header parts
		token := headerParts[1]

		// validate the token to make sure it is in sensible format
		v := validator.New()
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)

			return
		}

		// retrieve the details of the user associated with the authentication token
		// notice that ScopeAuthentication as the first parameter are being used
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}

			return
		}

		// add user information to the request context
		r = app.contextSetUser(r, user)

		// call the next handler in the chain
		next.ServeHTTP(w, r)
	})
}

// check if a user is not anonymous
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)

		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)

			return
		}

		next.ServeHTTP(w, r)
	})
}

// accept and return http.HandlerFunc to be used as wrapper for /v1/movies** endpoint
// check that a user is both authenticated and activated
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func (w http.ResponseWriter, r *http.Request) {
		// retrieve user information from the request context
		user := app.contextGetUser(r)

		// if user is not activated
		if !user.Activated {
			app.inactiveAccountResponse(w, r)

			return
		}

		// call the next handler in the chain
		next.ServeHTTP(w, r)
	})

	return app.requireAuthenticatedUser(fn)
}

func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// retrieve the user from request context
		user := app.contextGetUser(r)

		// get the slice codes of permissions for the user
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)

			return
		}

		// check if the slice includes the required permission
		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)

			return
		}

		// otherwise means the user has the required permission then
		// call the next handler in the chain
		next.ServeHTTP(w, r)
	}

	// wrap again with requireActivatedUser
	return app.requireActivatedUser(fn)
}

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// warn any caches that the response may be different
		w.Header().Add("Vary", "Origin")

		// Request-Method
		w.Header().Add("Vary", "Access-Control-Request-Method")

		// get the value of the request's Origin header
		origin := r.Header.Get("Origin")

		// check if there's an Origin request header present AND least one
		// trusted origin is configured
		if origin != "" && len(app.config.cors.trustedOrigins) != 0 {
			// loop and check if the request origin matches trusted list
			for i := range app.config.cors.trustedOrigins {
				if origin == app.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					// if the request has the HTTP method OPTIONS and contains
					// the "Access-Control-Request-Method" header treat it as
					// a pre-flight request
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						// write the headers along with a 200 OK status and return from the middleware
						// with no further action
						w.WriteHeader(http.StatusOK)

						return
					}
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) metrics(next http.Handler) http.Handler {
	// initialize the new expvar variables when the middleware chain is first built
	totalRequestsReceived := expvar.NewInt("total_requests_received")
	totalResponsesSent := expvar.NewInt("total_responses_sent")
	totalProcessingTimeMicroseconds := expvar.NewInt("total_processing_time_microseconds")

	// map to hold the count of responses for each HTTP status code
	totalResponsesSentByStatus := expvar.NewMap("total_responses_sent_by_status")

	// run the following code for every request
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// use the Add() method to the number of requests received by 1
		totalRequestsReceived.Add(1)

		// returns the metrics struct { Code int Duration time.Duration Written int64}
		metrics := httpsnoop.CaptureMetrics(next, w, r)

		// on the way back up the middleware chain, increment the number of responses
		totalResponsesSent.Add(1)

		// calculate the number of microseconds since the beginning of processing request,
		// then increment the total processing time by this amount
		totalProcessingTimeMicroseconds.Add(metrics.Duration.Microseconds())

		// the expvar map is string-keyed, use the strconv.Itoa()
		// function to convert the status code (in int) to a string
		totalResponsesSentByStatus.Add(strconv.Itoa(metrics.Code), 1)
	})
}
