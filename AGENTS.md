# AGENTS — Guide for automated coding agents

This document points an AI coding agent to the precise conventions, workflows and integration points that make contributors productive in this repository.

1) Big picture (what to know first)
- Clean Architecture: code is split by modules under `internal/modules/*` following Entities → UseCase → Repositories → Controllers. See `documentation/ARCHITECTURE.md` for details.
- Single binary API entry: `cmd/api/main.go` builds the HTTP server and wires `internal/config.NewApplication` (see `internal/config/app.go`) which composes modules, workers (asynq), WebSocket manager, TUS uploads, SSO providers and Casbin enforcer.

2) Key directories and responsibilities
- `cmd/api/` — application entrypoint, graceful shutdown and optional pprof (`PPROF_ENABLED` in `.env.example`).
- `internal/config/` — app wiring (DB, Redis, Enforcer, Storage provider). Inspect `app.go` to learn shared singletons and lifecycle.
- `internal/modules/*` — domain modules: usecases, repositories, controllers. New features should live here.
- `pkg/` — infrastructural helpers (jwt, sse, ws, storage, tus, telemetry). Prefer using these providers instead of reimplementing infra.
- `db/migrations`, `db/seeds` — database schema and seeders.
- `tests/` — unit/integration/e2e test suites; follow the repo's testing tags (`-tags=integration`, `-tags=e2e`).
- `web/` — Next.js frontend (separate repo-like directory).

3) Important patterns and project-specific conventions (do not violate)
- No passing full AppConfig into UseCases. Constructors should receive only the primitive values they need (see `documentation/ARCHITECTURE.md` under "Avoiding Circular Dependencies").
- Storage Provider: `pkg/storage.Provider` methods accept `context.Context` — always propagate context for deadlines/traces (mentioned in `ARCHITECTURE.md`).
- Casbin modifications that are part of DB transactions must use the `TransactionalEnforcer` wrapper (see `internal/config/app.go` where `permission.NewTransactionalEnforcer` is used).
- Tests: use the Singleton Container pattern for integration tests to avoid spinning containers per test run (see `documentation/guides/TESTING.md`). Integration/E2E tests require Docker.
- Background workers use `hibiken/asynq` (Redis); the project wires TaskDistributor and TaskProcessor in `internal/config/app.go` and starts them in `cmd/api/main.go`.

4) Common developer workflows (concrete commands)
- Start local infra (MySQL, Redis) and dev app (recommended):
  - docker compose -f docker-compose.dev.yml up --build
  - or `make docker-dev`
- Run the API locally (after `cp .env.example .env` and filling secrets):
  - `make run` (generates swagger then runs `cmd/api/main.go`)
  - `make build` builds the binary
- Tests:
  - Unit: `make test` / `make test-unit`
  - Integration: `make test-integration` (requires Docker)
  - E2E: `make test-e2e` (requires Docker)
  - Full: `make test-all`
- Generate docs: `make docs` (uses `swag` via pinned go run invocation in `Makefile`).
- Generate mocks: `make mocks` (project uses `mockery`, see `Makefile`).

5) Environment and secrets
- Copy `.env.example` to `.env` and set critical secrets: `JWT_ACCESS_SECRET`, `JWT_REFRESH_SECRET`, DB and Redis credentials. Many behaviors are feature-toggled via env flags (e.g. `WEBSOCKET_DISTRIBUTED_ENABLED`, `CASBIN_WATCHER_ENABLED`, `OTEL_ENABLED`).

6) Integration points & dependencies to be aware of
- MySQL (GORM) — DB migrations live in `db/migrations`. Migration commands are in `Makefile` (e.g. `make migrate-up`).
- Redis — session store, asynq backend, and WebSocket presence/pubsub (see `internal/config/app.go`).
- Casbin (DB-backed policies) — enforcer initialized in `internal/config`. When modifying policy state, prefer Transactional Enforcer patterns.
- Storage: local disk or S3-compatible (MinIO, R2). S3 settings are controlled in `.env.example` and `internal/config/storage.go`.
- TUS resumable uploads — registered hooks in `internal/config/app.go` (example: avatar hook registration).
- SSO providers (Google, Microsoft, GitHub) — `internal/config/app.go` shows how providers are registered.

7) Quick examples for agents
- To add a new module scaffold: `make gen-module` runs `cmd/gen/main.go` which produces standard module files under `internal/modules/<name>`.
- To run integration tests with a specific tag locally:
  - `GOTEST='go test -v ./tests/integration/... -tags=integration -p 1 -timeout=10m'` (Makefile target `test-integration` does this).
- To inspect runtime wiring, open `internal/config/app.go` and search for `NewApplication` — it shows how middlewares, enforcer, wsManager, sseManager and task processors are composed and started.

8) Files to read first (prioritized)
- `README.md` — project overview and env variables.
- `Makefile` — canonical build/test/dev commands.
- `internal/config/app.go` — wiring and lifecycle (most important for runtime behavior).
- `documentation/ARCHITECTURE.md` and `documentation/guides/TESTING.md` — architecture and test patterns.

9) Where to add agent assets / manifests
- Automated agent data in this repository (if used) is under `_bmad/_config/agents/` and related manifests in `_bmad/_config/manifest.yaml`. Follow the existing manifests when adding machine-readable agent definitions.

If something in this guide looks incomplete or you want me to expand a particular section (examples for writing UseCases, a checklist for adding Casbin policies, or a runnable dev container), say which section to expand and I will update `AGENTS.md`.

