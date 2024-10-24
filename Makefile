include .env

.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up 
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database "mysql://"${DSN} up

# db/migrations/drop: drop database
.PHONY: db/migrations/drop
db/migrations/drop: confirm
	@echo 'Running up migrations...'
	migrate -path ./migrations -database "mysql://"${DSN} drop