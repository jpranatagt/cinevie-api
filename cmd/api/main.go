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

	"api.cinevie.jpranata.tech/internal/data"

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
		maxOpenConns int
    maxIdleConns int
    maxIdleTime string
  }
}

// application dependencies
type application struct {
  config config
  logger *log.Logger
	models data.Models
}

func main() {
  // instance of config
  var cfg config

  // read the value for port and env from command-line flags into the config struct
  // default to port 4000 and development env
  flag.IntVar(&cfg.port, "port", 4000, "API server port.")
  flag.StringVar(&cfg.env, "env", "development", "Environment (development | staging | production).")
	// read the db-dsn commandline into the config struct use third argument as default db-dsn
  flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("CINEVIE_DB_DSN"), "PostgreSQL DSN")
	// notice the default value
  flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections.")
  flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections.")
  flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max connection idle time.")

  flag.Parse()

	 // logger prefixed with current date and time
  logger := log.New(os.Stdout, "", log.Ldate | log.Ltime)

	// openDB() creating connection pool
  db, err := openDB(cfg)
  if err != nil {
    logger.Fatal(err)
  }

  // defer, so connection closed before main() exits
  defer db.Close()

	// initialize Models struct passing in the connection pool as parameter
  logger.Printf("database connection pool established")
  app := &application {
    config: cfg,
    logger: logger,
		models: data.NewModels(db),
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

	// Set the maximum number of open (in-use + idle) connections in the pool. Note that
	// passing a value less than or equal to 0 will mean there is no limit.
	db.SetMaxOpenConns(cfg.db.maxOpenConns) // open = in-use + idle
	// Set the maximum number of idle connections in the pool. Again, passing a value
	// less than or equal to 0 will mean there is no limit.
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	// Use the time.ParseDuration() function to convert the idle timeout duration string
	// to a time.Duration type.

	// Use the time.ParseDuration() function to convert the idle timeout duration string
	// to a time.Duration type. return err if inputted time is in wrong format
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}
	// Set the maximum idle timeout.
	db.SetConnMaxIdleTime(duration)

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
