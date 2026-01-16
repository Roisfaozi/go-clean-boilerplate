# Go Clean Boilerplate - Gemini Context

This `GEMINI.md` file provides essential context for the Gemini agent to understand the structure, conventions, and operational procedures of this Go Clean Boilerplate project.

## Project Overview

**Name:** Go Clean Boilerplate - Enterprise Modular REST API
**Purpose:** An enterprise-ready Go boilerplate implementing Clean Architecture, RBAC with Casbin, Modular Audit Logging, and Distributed WebSocket scaling.
**Core Technologies:**
*   **Language:** Go 1.25.5+
*   **Web Framework:** Gin
*   **Database:** MySQL 8.0 (GORM)
*   **Caching/Session:** Redis 7
*   **Auth/Authz:** JWT & Casbin (RBAC)
*   **Infrastructure:** Docker, Docker Compose

## Architecture & Structure

The project strictly follows **Clean Architecture** principles.

*   **`cmd/api/`**: Application entry point (`main.go`).
*   **`internal/`**: Private application code.
    *   **`config/`**: Configuration and dependency injection.
    *   **`middleware/`**: HTTP middlewares (Auth, Casbin, CORS).
    *   **`modules/`**: Domain-specific modules (e.g., `auth`, `user`, `role`, `access`). Each module contains:
        *   `delivery/`: HTTP Handlers (Controllers).
        *   `usecase/`: Business Logic.
        *   `repository/`: Data Access (DB/Redis).
        *   `model/`: DTOs.
        *   `entity/`: Database entities.
    *   **`router/`**: Route registration.
*   **`pkg/`**: Reusable packages (e.g., `jwt`, `response`, `querybuilder`, `ws`).
*   **`db/`**: Database scripts (`migrations`, `seeds`).
*   **`tests/`**: Integration and E2E tests.

## Development Workflow & Commands

The project uses `Make` for automation. Always prefer `make` commands over direct `go` commands.

### Build & Run
*   **Run (Dev):** `make run` (or `air` for hot reload if installed).
*   **Build:** `make build` (Output: `main.exe` or `main`).
*   **Docker Dev:** `make docker-dev` (Starts MySQL & Redis).

### Database Management
*   **Migrate Up:** `make migrate-up`
*   **Migrate Down:** `make migrate-down`
*   **Create Migration:** `make migrate-create name=<migration_name>`
*   **Seed Data:** `make seed-up`

### Testing
*   **Unit Tests:** `make test` (Fast, mocks only).
*   **Integration Tests:** `make test-integration` (Requires Docker).
*   **E2E Tests:** `make test-e2e` (Requires Docker).
*   **All Tests:** `make test-all`
*   **Generate Mocks:** `make mocks` (Uses `mockery`).

### Code Quality
*   **Lint:** `make lint` (Uses `golangci-lint`).
*   **Format:** Standard `go fmt`.
*   **Docs:** `make docs` (Generates Swagger docs).

## Key Conventions

1.  **Dependency Injection:** Dependencies are injected into UseCases and Controllers (usually in `internal/config/app.go` or module initialization).
2.  **Error Handling:** Use custom errors from `pkg/exception`. Wrap errors for context.
3.  **Configuration:** All config is loaded from `.env` via `internal/config/config.go`.
4.  **Casbin:** Authorization policies are stored in the DB. Use `CasbinMiddleware` for route protection.
5.  **Dynamic Search:** Use `POST /search` endpoints with `pkg/querybuilder` for complex filtering.

## Environment Variables
Refer to `.env.example` for required variables. Key variables include:
*   `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
*   `REDIS_ADDR`, `REDIS_PASSWORD`
*   `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`
*   `CASBIN_ENABLED` (Default: `false`)

## Documentation
Additional detailed documentation is located in the `documentation/` directory, including:
*   `PROJECT_GUIDE.md` (Comprehensive guide)
*   `API_ACCESS_WORKFLOW.md` (RBAC details)
*   `TESTING_STRATEGY.md`
