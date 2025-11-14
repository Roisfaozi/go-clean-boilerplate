# Go parameters
GOCMD=go
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod

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
	$(SWAG_CLI) init -g cmd/api/main.go

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

# Clean up build artifacts and generated files
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@if (Test-Path $(BINARY_NAME)) { Remove-Item -Force $(BINARY_NAME) }
	@if (Test-Path ./docs) { Remove-Item -Recurse -Force ./docs }
	$(GOCLEAN)

