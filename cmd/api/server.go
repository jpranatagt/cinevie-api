package main

import (
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
    app.logger.PrintInfo("caught signal", map[string]string {
      "signal": s.String(),
    })

    // exit the application with 0 (success) status code
    os.Exit(0)
  }()

  // start the http server
  // INFO level with properties
  app.logger.PrintInfo("starting server", map[string]string{
    "addr": srv.Addr,
    "env": app.config.env,
  })

  return srv.ListenAndServe() // since db err has been declared above
}
