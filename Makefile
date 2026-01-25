include .env
export

MIGRATIONS_DIR := ./migrations

check-database-dsn:
ifndef DATABASE_URI
	$(error DATABASE_URI is not set. Please create .env file or set environment variable)
endif

build:
	cd cmd/gophermart && go build -o gophermart *.go

test:
	go test -v ./...

test-cov:
	go test -v ./... -coverprofile=coverage.out -coverpkg=./...
	go tool cover -func=coverage.out | grep "total:"

show-test-cov:
	go tool cover -html=coverage.out

run:
	go run cmd/gophermart/*.go

migrate-up: check-database-dsn
	@echo "Applying migrations"
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URI)" up

migrate-down: check-database-dsn
	@echo "Rolling back last migration..."
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URI)" down 1

migrate-version: check-database-dsn
	@echo "Current migration version:"
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URI)" version

migrate-create: check-database-dsn
ifndef NAME
	$(error NAME is required. Usage: make migrate-create NAME=migration_name)
endif
	@echo "Creating migration: $(NAME)"
	migrate create -ext sql -dir $(MIGRATIONS_DIR) $(NAME)

gen-json:
	cd internal/transport/http && easyjson -all dto/url.go

.PHONY: build test run