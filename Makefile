# Go parameters
GOCMD=go
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod

# Database
# Migration variables
MIGRATIONS_DIR = ./db/migrations
DB_DRIVER = mysql
DB_USER = root
DB_PASSWORD = Password0!
DB_PASSWORD_PROD =
DB_HOST = localhost
DB_HOST_PROD =
DB_PORT = 3307
DB_PORT_PROD = 3307
DB_NAME = gin_starter
DB_NAME_PROD =
DB_URL = "$(DB_DRIVER)://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)"
DB_URL_PROD = "$(DB_DRIVER)://$(DB_USER):$(DB_PASSWORD_PROD)@tcp($(DB_HOST_PROD):$(DB_PORT_PROD))/$(DB_NAME_PROD)"
DB_URL_STAG = "$(DB_DRIVER)://$(DB_USER):$(DB_PASSWORD_PROD)@tcp($(DB_HOST_PROD):$(DB_PORT_PROD))/$(DB_NAME)"


# Swagger CLI
SWAG_CLI=swag

# Binary name
BINARY_NAME=casbin-api.exe

# Default target executed when you just run `make`
.PHONY: all
all: help

# Displays help message
.PHONY: help
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  run          - Generate docs and run the main application."
	@echo "  build        - Build the application binary (output: $(BINARY_NAME))."
	@echo "  test         - Run all tests."
	@echo "  docs         - Generate Swagger/OpenAPI documentation."
	@echo "  tidy         - Tidy go.mod and go.sum files."
	@echo "  clean        - Remove build artifacts and generated documentation."
	@echo "  lint         - Run the static analysis linter (requires golangci-lint)."


# Generate docs and run the application
.PHONY: run
run: docs
	@echo "Running the application..."
	$(GORUN) ./cmd/api/main.go

# Build the application binary for production
.PHONY: build
build:
	@echo "Building the application binary..."
	$(GOBUILD) -o $(BINARY_NAME) ./cmd/api/main.go

# Run all tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v ./...

# Generate Swagger/OpenAPI documentation
.PHONY: docs
docs:
	@echo "Generating Swagger/OpenAPI documentation..."
	$(SWAG_CLI) init -g cmd/api/main.go --parseDependency --parseInternal --parseDepth 1

# Tidy go.mod and go.sum files
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

# Run linter (requires golangci-lint)
.PHONY: lint
lint:
	@echo "Running linter..."
	@if (-not (Get-Command golangci-lint -ErrorAction SilentlyContinue)) { \
		echo "golangci-lint is not installed. Please install it: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	}
	golangci-lint run

# Generate mocks
.PHONY: mocks
mocks:
	@echo "Generating mocks using .mockery.yaml..."
	@mockery

# Clean up build artifacts and generated files
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@if (Test-Path $(BINARY_NAME)) { Remove-Item -Force $(BINARY_NAME) }
	@if (Test-Path ./docs) { Remove-Item -Recurse -Force ./docs }
	$(GOCLEAN)


# Migration commands
.PHONY: migrate-install
migrate-install: ## Install golang-migrate
	@echo "Installing golang-migrate..."
	@go install -tags '$(DB_DRIVER)' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

.PHONY: migrate-create
migrate-create: ## Create new migration file (e.g., make migrate-create name=create_users_table)
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up: ## Run all up migrations
	@migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) -verbose up


.PHONY: migrate-up-1
migrate-up-1: ## Runcd  the next up migration
	@migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) -verbose up 1


.PHONY: migrate-down
migrate-down: ## Roll back all migrations
	@migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) -verbose down


.PHONY: migrate-down-1
migrate-down-1: ## Roll back the most recent migration
	@migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) -verbose down 1


.PHONY: migrate-force
migrate-force: ## Force a specific migration version (e.g., make migrate-force version=1)
	@migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) -verbose force $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-version
migrate-version: ## Show current migration version
	@migrate -path $(MIGRATIONS_DIR) -database $(DB_URL) version

# Seed commands
.PHONY: seed-up
seed-up: ## Seed initial data into the database (Go script)
	@echo "Seeding initial data using Go script..."
	@go run db/seeds/main.go

.PHONY: seed-down
seed-down: ## Rollback seeded data (if applicable, be careful!)
	@echo "Rolling back seeded data..."
	# This part needs to be carefully crafted based on your seed script's content
	# For 01_bootstrap.sql, it's not easily reversible without knowing generated UUIDs.
	# It's usually better to just re-seed in test environments after clean-up.
	@echo "Manual rollback may be required for complex seed data."


.PHONY: gemini
gemini: ## Set MySQL environment variables
	@powershell -ExecutionPolicy Bypass -Command "$$env:MYSQL_HOST='$(DB_HOST)'; $$env:MYSQL_PORT='$(DB_PORT)'; $$env:MYSQL_DATABASE='$(DB_NAME)'; $$env:MYSQL_USER='$(DB_USER)'; $$env:MYSQL_PASSWORD='$(DB_PASSWORD)'; gemini"

# -m gemini-2.5-pro