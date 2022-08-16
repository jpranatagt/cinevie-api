package main

import (
  "context"
  "database/sql"
  "flag"
  "os"
  "time"

	"api.cinevie.jpranata.tech/internal/data"
	"api.cinevie.jpranata.tech/internal/jsonlog"
	"api.cinevie.jpranata.tech/internal/mailer"

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

	// a new limiter struct containing fields for request-per-second, burst values,
  // and a boolean field to indicate that rate limiter is enabled or disabled
  limiter struct {
    rps float64
    burst int
    enabled bool
  }

	smtp struct {
    host string
    port int
    username string
    password string
    sender string
  }
}

// application dependencies
type application struct {
  config config
  logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
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

	// read rate limiter setting from command line
  flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum request per second.")
  flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst.")
  flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter.")

	// Read the SMTP server configuration settings into the config struct, using the
	// Mailtrap settings as the default values. IMPORTANT: If you're following along,
	// make sure to replace the default values for smtp-username and smtp-password
	// with your own Mailtrap credentials.
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "acdca05b068c66", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "d795c443f1a147", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@greenlight.jpranata.tech>", "SMTP sender")

  flag.Parse()

	// initialize a new jsonlog.Logger which writes any messages
  // *at or above* INFO severity level to standard out stream
  logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// openDB() creating connection pool
  db, err := openDB(cfg)
  if err != nil {
		// use PrintFatal() to write log at FATAL level and exit
    // no additional entry so pass nil
    logger.PrintFatal(err, nil)
  }

  // defer, so connection closed before main() exits
  defer db.Close()

	// initialize Models struct passing in the connection pool as parameter
	// INFO level
  logger.PrintInfo("database connection pool established", nil)
  app := &application {
    config: cfg,
    logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
  }

  err = app.serve()
	// print FATAL level and exit
	// fix panic: runtime error: invalid memory address or nil pointer dereference
	if err != nil {
  	logger.PrintFatal(err, nil)
	}
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
