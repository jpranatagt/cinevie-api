# include variables from the .envrc file
include .envrc


# ======================================================================== #
# HELPERS
# ======================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]


# ======================================================================== #
# DEVELOPMENT
# ======================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@go run ./cmd/api -db-dsn=${CINEVIE_DB_DSN}

## db/psql: connect to the database using psql and dsn from environment
.PHONY: db/psql
db/psql:
	@psql ${CINEVIE_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply the entire up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${CINEVIE_DB_DSN} up


# ======================================================================== #
# QUALITY CONTROL
# ======================================================================== #

## audit: tidy dependencies and format vet and test the entire code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies instead of using mirror proxy to host them
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor


# ======================================================================== #
# BUILD
# ======================================================================== #

current_time = $(shell date --iso-8601=seconds)
git_description = $(shell git describe --always --dirty --tags --long)
linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${git_description}'

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags=${linker_flags} -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/api ./cmd/api


# ======================================================================== #
# PRODUCTION
# ======================================================================== #

# fill with dedicated machine ip address or domain name
production_host_ip='api.cinevie.jpranata.tech'

# please use the same username between ./remote/setup/01.sh with
# the following production username for consistency and clarity
production_username='u_cinevie'

## production/setup: setup timezone, locales, user, universe repository, firewall, failban2, golang, postgresql database and caddy
.PHONY: production/setup
production/setup:
	rsync -rP --delete ./remote/setup root@${production_host_ip}:~
	ssh -t root@${production_host_ip} 'bash /root/setup/01.sh'
	sleep 10

## producion/connect: connect to the production server
.PHONY: production/connect
production/connect:
	ssh ${production_username}@${production_host_ip}

## production/deploy/api-starter: deploy api to production server as starter
.PHONY: production/deploy/api-starter
production/deploy/api-starter: production/connect
	rsync -rP --delete ./bin/linux_amd64/api ./migrations ${production_username}@${production_host_ip}:~
	ssh -t ${production_username}@${production_host_ip} 'migrate -path ~/migrations -database $$CINEVIE_DB_DSN up'

## production/deploy/api-update: deploy api to production server as update
.PHONY: production/deploy/api-update
production/deploy/api-update:
	rsync -rP --delete ./bin/linux_amd64/api ./migrations ${production_username}@${production_host_ip}:~
	ssh -t ${production_username}@${production_host_ip} 'migrate -path ~/migrations -database $$CINEVIE_DB_DSN up'

## production/configure/api.service: configure the production systemd api.service file
.PHONY: production/configure/api.service
production/configure/api.service:
	rsync -P ./remote/production/api.service ${production_username}@${production_host_ip}:~
	ssh -t ${production_username}@${production_host_ip} '\
		sudo mv ~/api.service /etc/systemd/system/ \
		&& sudo systemctl enable api \
		&& sudo systemctl restart api \
	'
## production/configure/caddyfile: configure the production of Caddyfile
.PHONY: production/configure/caddyfile
production/configure/caddyfile:
	rsync -P ./remote/production/Caddyfile ${production_username}@${production_host_ip}:~
	ssh -t ${production_username}@${production_host_ip} '\
		sudo mv ~/Caddyfile /etc/caddy/ \
		&& sudo systemctl reload caddy \
	'

## production/update: update binary and configuration on production machine
.PHONY: production/update
production/update: production/deploy/api-update production/configure/api.service production/configure/caddyfile

## production/all: run the entire setup and configurations for a ready to use production REST API
.PHONY: production/all
production/all: production/setup production/deploy/api-starter production/configure/api.service production/configure/caddyfile
