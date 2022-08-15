package main

import (
  "context"
  "database/sql"
  "flag"
  "fmt"
  "log"
  "net/http"
  "os"
  "time"

	// pq driver would register itself with database/sql
  // aliasing import to blank identifier(-) to stop compiler complaining
  // that the package not being used
  _ "github.com/lib/pq"
)

const version = "1.0.0"

// add a db field to hold configuration settings
// for now only holds DSN (Domain Source Name)
// from commandline flag
type config struct  {
  port int
  env string

	db struct {
    dsn string
  }
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
	// read the db-dsn commandline into the config struct use third argument as default db-dsn
  flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("CINEVI_DB_DSN"), "PostgreSQL DSN")

  flag.Parse()

	 // logger prefixed with current date and time
  logger := log.New(os.Stdout, "", log.Ldate | log.Ltime)

  // log a message that db connection pool has been successfully established
  logger.Printf("database connection pool established")

	// openDB() creating connection pool
  db, err := openDB(cfg)
  if err != nil {
    logger.Fatal(err)
  }

  // defer, so connection closed before main() exits
  defer db.Close()

  // log a message that db connection pool has been successfully established
  logger.Printf("database connection pool established")
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
    Handler:      app.routes(),
    IdleTimeout:  time.Minute,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 30 * time.Second,
  }

  // start the http server
  logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
  err = srv.ListenAndServe()
  logger.Fatal(err)
}

// return a sql.DB connection pool
func openDB(cfg config) (*sql.DB, error) {
  // create an empty connection
  db, err := sql.Open("postgres", cfg.db.dsn)
  if err != nil {
    return nil, err
  }

  // context with 5 seconds timeout deadline
  ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
  defer cancel()

  // use PingContext() to establish new connection to the database
  // if connection couldn't be established within 5 seconds deadline return error
  err = db.PingContext(ctx)
  if err != nil {
    return nil, err
  }

  // return the sql.DB connection pool and nil for the error
  return db, nil
}
