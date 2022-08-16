package main

import (
  "fmt"
	"net"
  "net/http"
	"sync"
	"time"

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
			// extract client IP address from request
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)

				return
			}

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
