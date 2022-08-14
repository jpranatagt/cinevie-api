package main

import (
  "flag"
  "fmt"
  "log"
  "net/http"
  "os"
  "time"
)

const version = "1.0.0"

type config struct  {
  port int
  env string
}

// dependencies holder
type application struct {
  config config
  logger *log.Logger
}

func main() {
  // instance of config
  var cfg config

  // read the value for port and env from command-line flags into the config struct
  // default to port 4000 and development env
  flag.IntVar(&cfg.port, "port", 4000, "API server port.")
  flag.StringVar(&cfg.env, "env", "development", "Environment (development | staging | production).")
  flag.Parse()

	 // logger prefixed with current date and time
  logger := log.New(os.Stdout, "", log.Ldate | log.Ltime)

  app := &application {
    config: cfg,
    logger: logger,
  }

  // servemux dispatch /v1/status route to statusHandler
  mux := http.NewServeMux()
  mux.HandleFunc("/v1/status", app.statusHandler)

  // http server with sensible timeout
  srv := &http.Server {
    Addr:         fmt.Sprintf(":%d", cfg.port),
    Handler:      mux,
    IdleTimeout:  time.Minute,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 30 * time.Second,
  }

  // start the http server
  logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
  err := srv.ListenAndServe()
  logger.Fatal(err)
}

