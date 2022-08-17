A REST API for retrieving and managing information about movies written in Go.

#### ENDPOINTS

Cinevie API would support following endpoints and actions:

| Method | URL Pattern               | Required Permission | Handler                          | Action                                     |
| ------ | ------------------------- | ------------------- | -------------------------------- | ------------------------------------------ |
| GET    | /v1/status                | -                   | statusHandler                    | Show application condition and information |
| POST   | /v1/movies                | movies:write        | createMovieHandler               | Create a new movie                         |
| GET    | /v1/movies/:id            | movies:read         | showMovieHandler                 | Show the details of a specific movie       |
| PATCH  | /v1/movies/:id            | movies:write        | updateMoviehandler               | Update the details of a specific movie     |
| DELETE | /v1/movies/:id            | movies:write        | deleteMovieHandler               | Delete a specific movie                    |
| GET    | /v1/movies                | movies:read         | listMovieHandler                 | Show the details of listed movies          |
| POST   | /v1/users                 | -                   | registerUserHandler              | Register a new user                        |
| PUT    | /v1/users/activated       | -                   | activateUserHandler              | Activate a specific user                   |
| POST   | /v1/tokens/activation     | -                   | createActivationTokenHandler     | Generate a new activation token            |
| POST   | /v1/tokens/authentication | -                   | createAuthenticationTokenHandler | Generate a new authentication token        |

#### DIRECTORY STRUCTURE

```
.
├── bin
├── cmd
│   └── api
│   └── main.go
├── go.mod
├── internal
├── Makefile
├── migrations
├── README.md
└── remote
```

**bin/** \
contain compiled application binaries for deployment

**cmd/api** \
application-specific code like running the server, reading and writing HTTP request, and managing authentication

**internal/** \
reusable code which imported by cmd/api (but not the other way around) for example database interaction, validation etc

**migrations/** \
SQL migrations files for database

**remote/** \
configuration files and setup scripts for production server

**go.mod** \
declare project dependencies, versions, and module path

**Makefile** \
common script for automating administrative tasks like building binaries and executing database migrations
