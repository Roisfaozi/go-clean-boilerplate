# Shell
SHELL := /bin/bash

# Go parameters
GOCMD=go
GORUN=$(GOCMD) run
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOMOD=$(GOCMD) mod
BENCHTIME=1s
BENCHMEM=true

# Worktree / local env
REPO_NAME := $(notdir $(CURDIR))
WORKTREE_ROOT ?= $(abspath $(CURDIR)/.worktrees)
WORKTREE_ROOT_FALLBACK ?= $(abspath $(CURDIR)/.worktrees)
WORKTREE_ROOT_SIBLING ?= $(abspath $(CURDIR)/../$(REPO_NAME)-worktrees)
WORKTREE_BRANCH ?= $(BRANCH)
WORKTREE_BASE ?= $(BASE)
WORKTREE_BRANCH_ARG ?= $(word 1,$(filter-out $@,$(MAKECMDGOALS)))
WORKTREE_BASE_ARG ?= $(word 2,$(filter-out $@,$(MAKECMDGOALS)))
WORKTREE_BRANCH_INPUT ?= $(strip $(WORKTREE_BRANCH)$(WORKTREE_BRANCH_ARG))
WORKTREE_BASE_INPUT ?= $(strip $(WORKTREE_BASE)$(WORKTREE_BASE_ARG))
ENV_LOCAL_FILE ?= .env.local
GO_VERIFY_PREFIX ?= PATH=/home/user/sdk/go/bin:$$PATH GOCACHE=/tmp/gocache
TEST_PKG ?= ./internal/... ./pkg/...

# Database
# Migration variables
MIGRATIONS_DIR = ./db/migrations
DB_DRIVER = mysql
DB_USER = root
DB_PASSWORD = Password0!
DB_PASSWORD_PROD = 
DB_HOST = localhost
DB_HOST_PROD = 
DB_PORT = 3306
DB_PORT_PROD = 3306
DB_NAME = gin_starter
DB_NAME_PROD = 
DB_URL = "$(DB_DRIVER)://$(DB_USER):$(DB_PASSWORD)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)"
DB_URL_PROD = "$(DB_DRIVER)://$(DB_USER):$(DB_PASSWORD_PROD)@tcp($(DB_HOST_PROD):$(DB_PORT_PROD))/$(DB_NAME_PROD)"
DB_URL_STAG = "$(DB_DRIVER)://$(DB_USER):$(DB_PASSWORD_PROD)@tcp($(DB_HOST_PROD):$(DB_PORT_PROD))/$(DB_NAME)"


# Swagger CLI (pin to module version to ensure consistent generated output)
SWAG_CLI=go run github.com/swaggo/swag/cmd/swag@v1.8.12

# Binary name
BINARY_NAME=main

# Default target executed when you just run `make`
.PHONY: all
all: help

# Displays help message
.PHONY: help
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  run               - Generate docs and run the main application."
	@echo "  build             - Build the application binary (output: $(BINARY_NAME))."
	@echo "  test              - Run unit tests only (fast, mocks only)."
	@echo "  test-unit         - Alias for test."
	@echo "  test-integration  - Run integration tests (requires Docker)."
	@echo "  test-e2e          - Run E2E tests (requires Docker)."
	@echo "  test-all          - Run all tests (unit + integration + e2e)."
	@echo "  test-coverage     - Run unit tests with coverage report."
	@echo "  test-coverage-all - Run all tests with coverage report."
	@echo "  test-race         - Run tests with race detector."
	@echo "  test-clean        - Clean test cache."
	@echo "  bench             - Run benchmarks."
	@echo "  docs              - Generate Swagger/OpenAPI documentation."
	@echo "  tidy              - Tidy go.mod and go.sum files."
	@echo "  clean             - Remove build artifacts and generated documentation."
	@echo "  lint              - Run the static analysis linter."
	@echo "  docker-dev        - Start the development environment with Docker Compose."
	@echo "  wt-new            - Create new git worktree from branch name, optional base."
	@echo "  wt-list           - List git worktrees."
	@echo "  wt-path           - Print worktree path for branch name."
	@echo "  wt-rm             - Remove worktree for branch name."
	@echo "  wt-prune          - Prune stale git worktree metadata."
	@echo "  wt-enter          - Print worktree path and ensure env is initialized."
	@echo "  env-init          - Create worktree-local .env.local with isolated ports."
	@echo "  env-sync          - Add missing keys from .env.example into .env.local."
	@echo "  dev-up            - Start worktree-local docker compose stack."
	@echo "  dev-down          - Stop worktree-local docker compose stack."
	@echo "  dev-reset         - Stop stack and remove worktree-local volumes."
	@echo "  dev-status        - Show current worktree/dev stack status."
	@echo "  migrate-up-local  - Run migrations against .env.local database."
	@echo "  migrate-down-local - Roll back migrations against .env.local database."
	@echo "  test-local        - Run narrow local tests, optional TEST_PKG=..."
	@echo "  doctor            - Check toolchain, env, worktree, and Docker readiness."
	@echo "  gen-module        - Generate boilerplate code for a new module."

.PHONY: wt-new
wt-new:
	branch_input="$(WORKTREE_BRANCH_INPUT)"; \
	base_input="$(WORKTREE_BASE_INPUT)"; \
	if [ -z "$$branch_input" ]; then \
		echo "Branch is required. Example: make wt-new feat/frontend-dashboard [dev|staging]"; \
		exit 1; \
	fi; \
	branch="$$branch_input"; \
	base="$$base_input"; \
	if [ -z "$$base" ]; then base="$$(git rev-parse --abbrev-ref HEAD)"; fi; \
	slug="$${branch//\//-}"; \
	path="$(WORKTREE_ROOT)/$$slug"; \
	if ! mkdir -p "$(WORKTREE_ROOT)" 2>/dev/null; then \
		if mkdir -p "$(WORKTREE_ROOT_FALLBACK)" 2>/dev/null; then \
			path="$(WORKTREE_ROOT_FALLBACK)/$$slug"; \
			WORKTREE_ROOT="$(WORKTREE_ROOT_FALLBACK)"; \
		elif mkdir -p "$(WORKTREE_ROOT_SIBLING)" 2>/dev/null; then \
			path="$(WORKTREE_ROOT_SIBLING)/$$slug"; \
			WORKTREE_ROOT="$(WORKTREE_ROOT_SIBLING)"; \
		else \
			echo "Cannot create worktree root inside repo or sibling fallback."; \
			exit 1; \
		fi; \
	fi; \
	if git worktree list --porcelain | grep -F "branch refs/heads/$$branch" >/dev/null; then \
		echo "Worktree for branch '$$branch' already exists."; \
		exit 1; \
	fi; \
	if [ -e "$$path" ]; then \
		echo "Target path already exists: $$path"; \
		exit 1; \
	fi; \
	echo "Creating worktree $$branch from $$base at $$path"; \
	git worktree add -b "$$branch" "$$path" "$$base"; \
	$(MAKE) -C "$$path" env-init >/dev/null; \
	$(MAKE) -C "$$path" env-sync >/dev/null; \
	echo "Worktree ready: $$path"; \
	echo "Next:"; \
	echo "  cd $$path"; \
	echo "  make dev-up"; \
	printf "%s\n" "$$path"

%:
	@:

.PHONY: wt-list
wt-list:
	@git worktree list

.PHONY: wt-path
wt-path:
	@if [ -z "$(WORKTREE_BRANCH_INPUT)" ]; then \
		echo "Branch is required. Example: make wt-path feat/frontend-dashboard"; \
		exit 1; \
	fi
	@branch="$(WORKTREE_BRANCH_INPUT)"; \
	path="$$(git worktree list --porcelain | awk -v branch="refs/heads/$$branch" 'BEGIN{p=""} /^worktree /{p=substr($$0,10)} /^branch / && $$2==branch {print p; exit}')"; \
	if [ -z "$$path" ]; then \
		echo "Worktree not found for branch '$$branch'"; \
		exit 1; \
	fi; \
	printf "%s\n" "$$path"

.PHONY: wt-rm
wt-rm:
	@if [ -z "$(WORKTREE_BRANCH_INPUT)" ]; then \
		echo "Branch is required. Example: make wt-rm feat/frontend-dashboard"; \
		exit 1; \
	fi
	@branch="$(WORKTREE_BRANCH_INPUT)"; \
	current_branch="$$(git rev-parse --abbrev-ref HEAD)"; \
	if [ "$$current_branch" = "$$branch" ]; then \
		echo "Cannot remove current active branch worktree: $$branch"; \
		exit 1; \
	fi; \
	path="$$(git worktree list --porcelain | awk -v branch="refs/heads/$$branch" 'BEGIN{p=""} /^worktree /{p=substr($$0,10)} /^branch / && $$2==branch {print p; exit}')"; \
	if [ -z "$$path" ]; then \
		echo "Worktree not found for branch '$$branch'"; \
		exit 1; \
	fi; \
	if [ -f "$$path/$(ENV_LOCAL_FILE)" ]; then \
		echo "Stopping local stack before removing worktree..."; \
		docker compose --env-file "$$path/$(ENV_LOCAL_FILE)" -f "$$path/docker-compose.dev.yml" down >/dev/null 2>&1 || true; \
	fi; \
	echo "Removing worktree $$branch at $$path"; \
	git worktree remove "$$path"

.PHONY: wt-enter
wt-enter:
	@if [ -z "$(WORKTREE_BRANCH_INPUT)" ]; then \
		echo "Branch is required. Example: make wt-enter feat/frontend-dashboard"; \
		exit 1; \
	fi
	@path="$$(git worktree list --porcelain | awk -v branch="refs/heads/$(WORKTREE_BRANCH_INPUT)" 'BEGIN{p=""} /^worktree /{p=substr($$0,10)} /^branch / && $$2==branch {print p; exit}')"; \
	if [ -z "$$path" ]; then \
		echo "Worktree not found for branch '$(WORKTREE_BRANCH_INPUT)'"; \
		exit 1; \
	fi; \
	$(MAKE) -C "$$path" env-init >/dev/null; \
	$(MAKE) -C "$$path" env-sync >/dev/null; \
	printf "%s\n" "$$path"

.PHONY: wt-prune
wt-prune:
	@git worktree prune

.PHONY: env-init
env-init:
	@slug="$$(git rev-parse --abbrev-ref HEAD 2>/dev/null || basename "$(CURDIR)")"; \
	slug="$${slug//\//-}"; \
	slug="$$(printf '%s' "$$slug" | tr '[:upper:]' '[:lower:]')"; \
	if [ ! -f "$(ENV_LOCAL_FILE)" ]; then cp .env.example "$(ENV_LOCAL_FILE)"; fi; \
	if [ ! -f "$(ENV_LOCAL_FILE)" ]; then echo "Failed to create $(ENV_LOCAL_FILE)"; exit 1; fi; \
	index="$$(printf '%s' "$$slug" | cksum | awk '{print ($$1 % 200) + 1}')"; \
	app_port="$$((8080 + index))"; \
	mysql_port="$$((3306 + index))"; \
	redis_port="$$((6379 + index))"; \
	jaeger_ui_port="$$((16686 + index))"; \
	otel_port="$$((4317 + index))"; \
	prom_port="$$((9090 + index))"; \
	grafana_port="$$((3000 + index))"; \
	mailpit_ui_port="$$((8025 + index))"; \
	mailpit_smtp_port="$$((1025 + index))"; \
	rustfs_port="$$((9000 + index))"; \
	rustfs_console_port="$$((9001 + index))"; \
	set_kv() { \
		key="$$1"; value="$$2"; file="$(ENV_LOCAL_FILE)"; \
		if grep -qE "^$${key}=" "$$file"; then \
			sed -i "s|^$${key}=.*|$${key}=$${value}|" "$$file"; \
		else \
			printf "%s=%s\n" "$$key" "$$value" >> "$$file"; \
		fi; \
	}; \
	set_kv WORKTREE_SLUG "$$slug"; \
	set_kv COMPOSE_PROJECT_NAME "$(REPO_NAME)-$$slug"; \
	set_kv APP_PORT "$$app_port"; \
	set_kv MYSQL_PORT "$$mysql_port"; \
	set_kv REDIS_PORT "$$redis_port"; \
	set_kv JAEGER_UI_PORT "$$jaeger_ui_port"; \
	set_kv OTEL_GRPC_PORT "$$otel_port"; \
	set_kv PROMETHEUS_PORT "$$prom_port"; \
	set_kv GRAFANA_PORT "$$grafana_port"; \
	set_kv MAILPIT_UI_PORT "$$mailpit_ui_port"; \
	set_kv MAILPIT_SMTP_PORT "$$mailpit_smtp_port"; \
	set_kv RUSTFS_PORT "$$rustfs_port"; \
	set_kv RUSTFS_CONSOLE_PORT "$$rustfs_console_port"; \
	set_kv REDIS_ADDR "localhost:$$redis_port"; \
	echo "Initialized $(ENV_LOCAL_FILE) for $$slug"; \
	echo "APP_PORT=$$app_port MYSQL_PORT=$$mysql_port REDIS_PORT=$$redis_port"

.PHONY: env-sync
env-sync:
	@if [ ! -f "$(ENV_LOCAL_FILE)" ]; then \
		echo "$(ENV_LOCAL_FILE) not found. Run 'make env-init' first."; \
		exit 1; \
	fi
	@while IFS= read -r line; do \
		case "$$line" in \
			''|'#'*) continue ;; \
		esac; \
		key="$${line%%=*}"; \
		if ! grep -qE "^$${key}=" "$(ENV_LOCAL_FILE)"; then \
			printf "%s\n" "$$line" >> "$(ENV_LOCAL_FILE)"; \
		fi; \
	done < .env.example; \
	echo "Synced missing keys into $(ENV_LOCAL_FILE)"

.PHONY: dev-up
dev-up:
	@if [ ! -f "$(ENV_LOCAL_FILE)" ]; then \
		echo "$(ENV_LOCAL_FILE) missing, auto-initializing..."; \
		$(MAKE) env-init >/dev/null; \
	fi
	@$(MAKE) env-sync >/dev/null; \
	@docker compose --env-file "$(ENV_LOCAL_FILE)" -f docker-compose.dev.yml up -d --build

.PHONY: dev-down
dev-down:
	@if [ ! -f "$(ENV_LOCAL_FILE)" ]; then \
		echo "$(ENV_LOCAL_FILE) not found. Nothing to stop."; \
		exit 0; \
	fi
	@docker compose --env-file "$(ENV_LOCAL_FILE)" -f docker-compose.dev.yml down

.PHONY: dev-reset
dev-reset:
	@if [ ! -f "$(ENV_LOCAL_FILE)" ]; then \
		echo "$(ENV_LOCAL_FILE) not found. Nothing to reset."; \
		exit 0; \
	fi
	@docker compose --env-file "$(ENV_LOCAL_FILE)" -f docker-compose.dev.yml down -v --remove-orphans

.PHONY: dev-status
dev-status:
	@if [ ! -f "$(ENV_LOCAL_FILE)" ]; then \
		echo "$(ENV_LOCAL_FILE) missing, auto-initializing..."; \
		$(MAKE) env-init >/dev/null; \
	fi
	@branch="$$(git rev-parse --abbrev-ref HEAD)"; \
	slug="$$(grep '^WORKTREE_SLUG=' "$(ENV_LOCAL_FILE)" | tail -n1 | cut -d= -f2-)"; \
	project="$$(grep '^COMPOSE_PROJECT_NAME=' "$(ENV_LOCAL_FILE)" | tail -n1 | cut -d= -f2-)"; \
	app_port="$$(grep '^APP_PORT=' "$(ENV_LOCAL_FILE)" | tail -n1 | cut -d= -f2-)"; \
	mysql_port="$$(grep '^MYSQL_PORT=' "$(ENV_LOCAL_FILE)" | tail -n1 | cut -d= -f2-)"; \
	redis_port="$$(grep '^REDIS_PORT=' "$(ENV_LOCAL_FILE)" | tail -n1 | cut -d= -f2-)"; \
	echo "branch=$$branch"; \
	echo "slug=$$slug"; \
	echo "compose_project=$$project"; \
	echo "app_port=$$app_port mysql_port=$$mysql_port redis_port=$$redis_port"; \
	docker compose --env-file "$(ENV_LOCAL_FILE)" -f docker-compose.dev.yml ps

.PHONY: migrate-up-local
migrate-up-local:
	@if [ ! -f "$(ENV_LOCAL_FILE)" ]; then \
		echo "$(ENV_LOCAL_FILE) not found. Run 'make env-init' first."; \
		exit 1; \
	fi
	@set -a; source "$(ENV_LOCAL_FILE)"; set +a; \
	dsn="mysql://$$MYSQL_USER:$$MYSQL_PASSWORD@tcp(localhost:$$MYSQL_PORT)/$$MYSQL_DBNAME"; \
	migrate -path $(MIGRATIONS_DIR) -database "$$dsn" -verbose up

.PHONY: migrate-down-local
migrate-down-local:
	@if [ ! -f "$(ENV_LOCAL_FILE)" ]; then \
		echo "$(ENV_LOCAL_FILE) not found. Run 'make env-init' first."; \
		exit 1; \
	fi
	@set -a; source "$(ENV_LOCAL_FILE)"; set +a; \
	dsn="mysql://$$MYSQL_USER:$$MYSQL_PASSWORD@tcp(localhost:$$MYSQL_PORT)/$$MYSQL_DBNAME"; \
	migrate -path $(MIGRATIONS_DIR) -database "$$dsn" -verbose down 1

.PHONY: test-local
test-local:
	@echo "Running local tests: $(TEST_PKG)"
	@$(GO_VERIFY_PREFIX) $(GOTEST) -v $(TEST_PKG)

.PHONY: doctor
doctor:
	@echo "branch=$$(git rev-parse --abbrev-ref HEAD)"
	@echo "worktree_root=$(WORKTREE_ROOT)"
	@echo "worktree_root_fallback=$(WORKTREE_ROOT_FALLBACK)"
	@echo "worktree_root_sibling=$(WORKTREE_ROOT_SIBLING)"
	@command -v git >/dev/null && echo "git=ok" || echo "git=missing"
	@command -v docker >/dev/null && echo "docker=ok" || echo "docker=missing"
	@command -v pnpm >/dev/null && echo "pnpm=ok" || echo "pnpm=missing"
	@PATH=/home/user/sdk/go/bin:$$PATH command -v go >/dev/null && echo "go=ok" || echo "go=missing"
	@if [ -f "$(ENV_LOCAL_FILE)" ]; then \
		echo "env_local=present"; \
		grep -E '^(WORKTREE_SLUG|COMPOSE_PROJECT_NAME|APP_PORT|MYSQL_PORT|REDIS_PORT)=' "$(ENV_LOCAL_FILE)"; \
	else \
		echo "env_local=missing"; \
	fi


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

# --- TESTING ---

define run_go_test
	@set -o pipefail; \
	output=$$(mktemp); \
	$(GO_VERIFY_PREFIX) $(GOTEST) $(1) >$$output 2>&1; \
	status=$$?; \
	if [ $$status -eq 0 ]; then \
		echo "PASS"; \
	else \
		echo "FAIL"; \
		awk '/^=== RUN / || /^--- FAIL: / || /^FAIL([[:space:]]|$$)/ || /^panic: / {print}' "$$output"; \
	fi; \
	rm -f "$$output"; \
	exit $$status
endef

# Run unit tests only (exclude integration and e2e folders)
.PHONY: test
test:
	@echo "Running unit tests (internal and pkg)..."
	$(call run_go_test,-p 1 -parallel 1 -count=1 -v ./internal/... ./pkg/...) 

.PHONY: test-unit
test-unit: test

# Run integration tests (requires Docker)
.PHONY: test-integration
test-integration:
	@echo "Running integration tests..."
	$(call run_go_test,-p 1 -parallel 1 -count=1 -v ./tests/integration/... -tags=integration -timeout=10m)

# Run E2E tests (requires Docker)
.PHONY: test-e2e
test-e2e:
	@echo "Running E2E tests..."
	$(call run_go_test,-p 1 -parallel 1 -count=1 -v ./tests/e2e/... -tags=e2e -timeout=15m)

# Run all tests (unit + integration + e2e)
.PHONY: test-all
test-all:
	@set +e; \
	$(MAKE) test; status1=$$?; \
	$(MAKE) test-integration; status2=$$?; \
	$(MAKE) test-e2e; status3=$$?; \
	if [ $$status1 -eq 0 ] && [ $$status2 -eq 0 ] && [ $$status3 -eq 0 ]; then \
		echo "OVERALL: PASS"; \
	else \
		echo "OVERALL: FAIL"; \
	fi; \
	exit 0

# Run unit tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running unit tests coverage..."
	@set -o pipefail; \
	output=$$(mktemp); \
	$(GO_VERIFY_PREFIX) $(GOTEST) -p 1 -parallel 1 -count=1 -coverprofile=coverage_unit.out -covermode=atomic -v ./internal/... ./pkg/... >$$output 2>&1; \
	status=$$?; \
	if [ $$status -eq 0 ]; then \
		$(GOCMD) tool cover -html=coverage_unit.out -o coverage_unit.html; \
		echo "PASS"; \
		echo "Unit test coverage report: coverage_unit.html"; \
	else \
		echo "FAIL"; \
		awk '/^=== RUN / || /^--- FAIL: / || /^FAIL([[:space:]]|$$)/ || /^panic: / {print}' "$$output"; \
	fi; \
	rm -f "$$output"; \
	exit $$status

# Run all tests with coverage (Sequential to avoid singleton DB race conditions)
.PHONY: test-coverage-all
test-coverage-all:
	@echo "Running all tests coverage..."
	@set -o pipefail; \
	output=$$(mktemp); \
	$(GO_VERIFY_PREFIX) $(GOTEST) -p 1 -parallel 1 -count=1 -coverprofile=coverage_all.out -covermode=atomic -v ./... -tags=integration,e2e -timeout=20m >$$output 2>&1; \
	status=$$?; \
	if [ $$status -eq 0 ]; then \
		$(GOCMD) tool cover -html=coverage_all.out -o coverage_all.html; \
		echo "PASS"; \
		echo "Full coverage report: coverage_all.html"; \
	else \
		echo "FAIL"; \
		awk '/^=== RUN / || /^--- FAIL: / || /^FAIL([[:space:]]|$$)/ || /^panic: / {print}' "$$output"; \
	fi; \
	rm -f "$$output"; \
	exit $$status

# Run tests with race detector
.PHONY: test-race
test-race:
	$(call run_go_test,-race -v ./...)

# Clean test cache
.PHONY: test-clean
test-clean:
	$(GOCLEAN) -testcache

# --- BENCHMARKING ---

# Run benchmarks
.PHONY: bench
bench:
	$(GOTEST) -run=^$$ -bench=. -benchmem -benchtime=$(BENCHTIME) -count=5 ./...

# Run benchmarks with race detector
.PHONY: bench-race
bench-race:
	$(GOTEST) -race -run=^$$ -bench=. -benchmem -benchtime=$(BENCHTIME) -count=5 ./...

# --- DOCKER ---

.PHONY: docker-dev
docker-dev: ## Start development environment
	@echo "Starting development environment..."
	@if [ -f "$(ENV_LOCAL_FILE)" ]; then \
		docker compose --env-file "$(ENV_LOCAL_FILE)" -f docker-compose.dev.yml up -d --build; \
	else \
		docker compose -f docker-compose.dev.yml up -d --build; \
	fi

.PHONY: docker-prod
docker-prod: ## Start production environment (initial)
	@echo "Starting production environment (initial setup)..."
	docker compose -f docker-compose.prod.yml up -d --build

.PHONY: docker-down
docker-down: ## Stop all containers
	@echo "Stopping all containers..."
	@if [ -f "$(ENV_LOCAL_FILE)" ]; then \
		docker compose --env-file "$(ENV_LOCAL_FILE)" -f docker-compose.dev.yml down; \
	else \
		docker compose -f docker-compose.dev.yml down; \
	fi
	docker compose -f docker-compose.prod.yml down

.PHONY: deploy
deploy: ## Run Blue-Green deployment
	@echo "Running Blue-Green deployment..."
	@bash deploy/scripts/deploy.sh

# --- TOOLS ---

# Generate Swagger/OpenAPI documentation
.PHONY: docs
docs:
	@echo "Generating Swagger/OpenAPI documentation..."
	$(SWAG_CLI) init -g cmd/api/main.go --parseDependency --parseInternal --parseDepth 1
	@echo "Swagger documentation generated successfully!"

# Tidy go.mod and go.sum files
.PHONY: tidy
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy

# Run linter (requires golangci-lint)
.PHONY: lint
lint:
	@echo "Running linter..."
	golangci-lint run

.PHONY: lint-fix
lint-fix:
	@echo "Running linter with auto-fix..."
	golangci-lint run --fix

.PHONY: vulcek
vulcek:
	@echo "Running vulnerability check with govulncheck ./..."
	govulncheck ./...

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
gemini: ## Set MySQL environment variables and run gemini using zsh
		@echo "Starting gemini with MySQL environment variables (zsh)..."
		@env MYSQL_HOST="$(DB_HOST)" MYSQL_PORT="$(DB_PORT)" MYSQL_DATABASE="$(DB_NAME)" MYSQL_USER="$(DB_USER)" MYSQL_PASSWORD="$(DB_PASSWORD)" zsh -c 'gemini'
	@#powershell -ExecutionPolicy Bypass -Command "$$env:MYSQL_HOST='$(DB_HOST)'; $$env:MYSQL_PORT='$(DB_PORT)'; $$env:MYSQL_DATABASE='$(DB_NAME)'; $$env:MYSQL_USER='$(DB_USER)'; $$env:MYSQL_PASSWORD='$(DB_PASSWORD)'; gemini"

# Generate new module boilerplate
.PHONY: gen-module
gen-module:
	@echo "Enter module name (e.g. product): "
	@read module; \
	$(GORUN) cmd/gen/main.go -name $$module
