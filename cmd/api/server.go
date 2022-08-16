package main

import (
	"context"
	"errors"
  "fmt"
  "log"
  "net/http"
  "os"
  "os/signal"
  "syscall"
  "time"
)

func (app *application) serve() error {
  // http server with sensible timeout
  srv := &http.Server {
    Addr:         fmt.Sprintf(":%d", app.config.port),
    Handler:      app.routes(), // use httprouter
    // Go log.logger instance with the log.New() function, passing
    // in custom logger as the first parameter. The "" and 0 indicate
    // that the log.Logger instance should not use a prefix or any flags
    ErrorLog:     log.New(app.logger, "", 0),
    IdleTimeout:  time.Minute,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 30 * time.Second,
  }

	// use this channel to receive any errors returned by graceful Shutdown()
  shutdownError := make(chan error)


  go func() {
    // create quit channel which carries os.Signal values
    // use buffer in case with size 1
    // because signal.Notify() would not wait for receiver
    // to be available when sending signal to quit channel
    quit := make(chan os.Signal, 1)

    // use signal.Notify() to listen for incoming SIGINT and SIGTERM signals
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

    // read the signal from the quit channel, this code will block
    // until signal is received
    s := <-quit

    // log signal has been caught with String() method
    // to inform signal name and include in log entry properties
	app.logger.PrintInfo("shutting down server", map[string]string {
      "signal": s.String(),
    })

	// 5 second timeout context
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

	// shutdown channel if it returns an error
	err := srv.Shutdown(ctx)
	if err != nil {
		shutdownError <- err
	}

	// waiting for any background go routines to complete their tasks
	app.logger.PrintInfo("completing background tasks", map[string]string {
		"addr": srv.Addr,
	})

	// call the Wait() to block until WaitGroup counter is zero
	app.wg.Wait()
	shutdownError <- nil
  }()

  // start the http server
  // INFO level with properties
  app.logger.PrintInfo("starting server", map[string]string{
    "addr": srv.Addr,
    "env": app.config.env,
  })

	// Calling Shutdown() on server will cause ListenAndServe() to immediately
	// return a http.ErrServerClosed error. So if this error happen, it is actually a
	// good thing and an indication that the graceful shutdown has started.
	// Check specifically for this, only returning the error if it is NOT http.ErrServerClosed.
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// Otherwise, we wait to receive the return value from Shutdown() on the
	// shutdownError channel. If return value is an error, we know that there was a
	// problem with the graceful shutdown and we return the error.
	err =<-shutdownError
	if err != nil {
		return err
	}
	// At this point we know that the graceful shutdown completed successfully and we
	// log a "stopped server" message.
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
