# Go Conventions

## Purpose

Guide for backend Go code structure, module boundaries, error handling, security-sensitive behavior, and testing expectations in this repo.

## Architectural shape

Primary backend layering in modules is:

- controller or delivery/http
- usecase
- repository
- model or entity

Shared runtime infrastructure lives under:

- `internal/config`
- `internal/middleware`
- `internal/worker`
- `pkg/*`

## Ownership rules

- controller
  - bind request
  - validate input
  - translate response or errors
- usecase
  - business rules
  - orchestration across repositories or services
  - transaction ownership when needed
- repository
  - DB and query details
  - GORM-specific operations
- app/config wiring
  - dependency construction and lifecycle

## Dependency and wiring rules

- inspect `internal/config/app.go` for real dependency composition before refactor
- do not pass full app config into usecases when only a few values or dependencies are needed
- avoid ad hoc globals when constructor or wiring already owns dependency lifecycle

## Middleware and boundary rules

- auth, session, tenant, API-key, user-status, and Casbin checks belong in middleware and usecase boundaries, not duplicated ad hoc in handlers
- route ownership and group semantics come from `internal/router/router.go`
- upload, realtime, worker, and webhook paths are special trust or side-effect boundaries; do not treat them like ordinary CRUD paths

## Error and response rules

- prefer existing project exception or response helpers where already used
- do not leak internal errors from auth, token, Casbin, or DB paths into user-facing messages
- preserve security behavior around timing mitigation, session validation, and status checks

## Security-sensitive code rules

- do not weaken `pkg/querybuilder` sensitive field denylist
- do not add protected endpoints without deciding API-key scope, JWT/session, tenant context, and Casbin route group
- do not bypass Redis-backed session validation by only parsing JWT
- treat WebSocket origin or ticket checks as security controls
- treat upload metadata and TUS hook dispatch as trust boundaries

## Formatting and linting

- follow standard Go formatting with `gofmt`
- `Makefile` exposes `make lint` and `make lint-fix` for `golangci-lint`
- CI runs lint before test or build jobs

## Testing expectations

- unit tests stay close to packages under `internal` and `pkg`
- integration and E2E tests live under `tests/` and often require Docker
- regenerate mocks with `make mocks` when interfaces change
- run narrow tests first, then integration/E2E if change touches route, DB, Redis, Casbin, worker, tenant, upload, or realtime boundaries

## Common mistakes

- putting business logic into controller because it is convenient
- bypassing constructor wiring with temporary globals or side channels
- making security boundary change inside usecase while route or middleware still implies old contract
- claiming backend-only change when frontend proxies or shared types actually depend on it
