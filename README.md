A REST API for retrieving and managing information about movies written in Go.

Cinevie API would support following endpoints and actions:

| Method | URL Pattern | Handler       | Action                                     |
| ------ | ----------- | ------------- | ------------------------------------------ |
| GET    | /v1/status  | statusHandler | Show application condition and information |

# DIRECTORY STRUCTURE

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

bin/
contain compiled application binaries for deployment

cmd/api
application-specific code like running the server, reading and writing HTTP request, and managing authentication

internal/
reusable code which imported by cmd/api (but not the other way around) for example database interaction, validation etc

migrations/
SQL migrations files for database

remote/
configuration files and setup scripts for production server

go.mod
declare project dependencies, versions, and module path

Makefile
common script for automating administrative tasks like building binaries and executing database migrations