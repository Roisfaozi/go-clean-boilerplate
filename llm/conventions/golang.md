# Go Conventions

## Architectural shape

- Put backend business features under `internal/modules/<module>`.
- Preserve module layering already used in repo: `delivery/http` -> `usecase` -> `repository` -> `entity/model`.
- Wire module-local dependencies in `internal/modules/*/module.go`.
- Wire cross-module dependencies in `internal/config/app.go`.
- Do not pass full `AppConfig` into usecases; pass only required primitive values and dependencies.
- Keep generated/scaffolded module shape compatible with `cmd/gen/main.go` and `make gen-module`.

## Dependency injection

- Constructors should take explicit dependencies such as repos, validator, logger, transaction manager, enforcer, Redis, storage, worker distributor, or managers.
- Controllers should not create GORM DB connections, Redis clients, enforcers, workers, or repositories directly.
- Cross-module calls should be represented as usecase/repository interfaces or constructor dependencies, not package-level globals.

## Context and transactions

- Propagate `context.Context` from HTTP request into usecases, repositories, storage, worker queueing, and audit paths where supported.
- Use transaction manager abstractions from `pkg/tx` for multi-write behavior.
- When Casbin changes must share transaction semantics with DB writes, use transactional enforcer flow.
- In tenant-aware paths, preserve org/user context; do not re-derive it from untrusted inputs if middleware already resolved it.

## Controllers and validation

- Controllers bind request data with Gin helpers such as `ShouldBindJSON`, path params, and query params.
- Controllers validate request structs with injected `validator.Validate` where module pattern uses it.
- Use `pkg/validation.FormatValidationErrors` when adjacent controllers already format validation messages that way.
- Controllers should delegate business rules to usecases.

## Error and response patterns

- Prefer project exception/response helpers where already used.
- Do not leak internal errors from auth, token, Casbin, or DB paths into user-facing messages.
- Keep auth/session/tenant/Casbin checks in middleware and usecase boundaries, not duplicated ad hoc in handlers.
- Preserve security behavior around timing mitigation, session validation, and status checks.

## Security-sensitive code

- Do not weaken `pkg/querybuilder` sensitive field denylist.
- Do not add protected endpoints without deciding API-key scope, JWT/session, tenant context, and Casbin route group.
- Do not bypass Redis-backed session validation by only parsing JWT.
- Treat WebSocket origin/ticket checks as security controls.
- Treat upload metadata and TUS hook dispatch as trust boundaries.

## Formatting and linting

- Follow Go formatting with `gofmt`/standard Go formatting.
- `Makefile` exposes `make lint` and `make lint-fix` for `golangci-lint`.
- CI runs `golangci-lint-action` before test/build jobs.

## Testing

- Unit tests stay close to packages under `internal` and `pkg`.
- Integration and E2E tests live under `tests/` and require Docker.
- Regenerate mocks with `make mocks` when interfaces change.
- Run narrow tests first, then integration/E2E if the change touches route, DB, Redis, Casbin, worker, tenant, upload, or realtime boundaries.
