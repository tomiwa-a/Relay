include .envrc

.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/ -db-dsn=${RELAY_DB_DSN}

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	psql ${RELAY_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database ${RELAY_DB_DSN} up

## db/migrations/down: rollback the last database migration
.PHONY: db/migrations/down
db/migrations/down: confirm
	@echo 'Rolling back last migration...'
	migrate -path ./migrations -database ${RELAY_DB_DSN} down 1

## db/migrations/reset: rollback all database migrations
.PHONY: db/migrations/reset
db/migrations/reset: confirm
	@echo 'Rolling back all migrations...'
	migrate -path ./migrations -database ${RELAY_DB_DSN} down

## docker/up: start all docker containers in background
.PHONY: docker/up
docker/up:
	docker-compose up -d

## docker/down: stop and remove all docker containers
.PHONY: docker/down
docker/down:
	docker-compose down

## sqlc: generate type-safe code from SQL
.PHONY: sqlc
sqlc:
	sqlc generate

## audit: tidy dependencies and format, vet and test all code
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...