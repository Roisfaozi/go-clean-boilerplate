# Backend Map

## Entrypoints

- API binary: `cmd/api/main.go` (boots `config.CoreFxModule` via `go.uber.org/fx` in `internal/config/fx.go`)
- Generator binary: `cmd/gen/main.go`
- Composition root: `internal/config/app.go` with extracted helper wiring in `internal/config/app_helpers.go`
- Router: `internal/router/http_router.go` (`SetupStdRouter`)

## Shared infrastructure

- config loading: `internal/config/config.go`
- DB: `internal/config/sqlx.go` (`NewSQLXDatabase` returning `*sqlx.DB`)
- Redis: `internal/config/redis.go`
- Casbin: `internal/config/casbin.go`, `internal/config/casbin_model.conf`, `pkg/casbinadapter`
- storage: `internal/config/storage.go`, `pkg/storage/*`
- transaction manager: `pkg/tx` (`SQLXTransactionManager`, `DBTX`, `DBTXFromContext`)
- SQL tenant/visibility helpers: `pkg/database` (`SQLOrganizationClause`, `SQLOrganizationVisibilityClause`)
- JWT: `pkg/jwt`
- SSE: `pkg/sse`
- WebSocket: `pkg/ws`
- TUS: `pkg/tus`
- telemetry: `pkg/telemetry`
- query builder: `pkg/querybuilder` (`BuildRawQuery`, `GetDBFieldName`)

Runtime wiring direction:

- `internal/config/app.go` owns final cross-module composition.
- `internal/config/app_helpers.go` owns extracted helper builders for modules, SSO providers, realtime broadcaster, and TUS handler setup.
- `internal/modules/*/module.go` owns module-local construction.
- controllers should not create repositories or global clients directly.
- usecases should receive explicit dependencies and primitives.

## Modules

- `access`: endpoint and access-right registry.
- `api_key`: API key issuance, revocation, cache-backed lookup, organization access checks.
- `audit`: audit logs, outbox, websocket/worker side effects.
- `auth`: register, login, refresh/logout, password reset, email verification, SSO, WebSocket ticket.
- `organization`: organization lifecycle, membership, invitation, tenant reader/cache.
- `permission`: Casbin policy operations, role/user permissions, access-right expansion.
- `project`: tenant-scoped project CRUD.
- `role`: role CRUD and role-policy cleanup through permission usecase.
- `stats`: dashboard stats and activity/insight summaries.
- `user`: user CRUD/profile/status/avatar, audit/webhook side effects.
- `webhook`: webhook config, logs, and asynchronous dispatch.

Module dependency hotspots:

- `auth` depends on user, organization, audit/event publishing, Casbin adapter, Redis token repository, worker, ticket manager, and SSO providers.
- `organization` depends on member/invitation repos, cached org reader, user repo, worker, enforcer, and presence reader.
- `permission` depends on role/user/access repos and audit usecase.
- `user` depends on auth, audit, webhook, storage, transaction manager, and enforcer.
- `role` depends on permission usecase for role-policy cleanup.

## Route ownership by module

- `auth`: `/auth/*` public and authenticated auth endpoints
- `organization`: `/organizations/*` across public, authenticated, tenant, and admin flows
- `user`: `/users/*` across public, authenticated, and authorized flows
- `stats`: `/stats/*` in authenticated scope
- `permission`: `/permissions/*` and `/permissions/check-batch`
- `access`: `/access-rights/*` and `/endpoints/*`
- `role`: `/roles/*`
- `project`: `/projects/*`
- `api_key`: `/api-keys/*`
- `audit`: `/audit-logs/*`
- `webhook`: `/webhooks/*`
- upload surface: `/api/v1/upload/files/*`

## Request path pattern

Common backend request path:

1. `http.ServeMux` route in `internal/router/http_router.go` (`SetupStdRouter`) wrapped with `delivery.Chain`.
2. Middleware applies request ID, logging, recovery, security headers, CORS, API-key, JWT/session, CSRF, user status, tenant, and/or Casbin checks.
3. Stdlib `HTTP*` controller method parses and validates request.
4. Usecase executes business behavior and transaction handling (`pkg/tx`).
5. Repository executes parameterized `sqlx` raw SQL queries via `tx.DBTX` (`DBTXFromContext(ctx)` or `*sqlx.DB`).
6. Side effects go through worker, audit, webhook, SSE, or WebSocket managers.

This pattern is visible in module constructors and route/controller/repository files, not only in documentation.

Controller pattern:

- bind JSON via `response.DecodeJSON` or read path/query params via `r.PathValue(...)` / `r.URL.Query()` (legacy Gin handler methods remain alongside `HTTP*` handlers during phased cutover)
- validate with injected validator where applicable
- read auth/tenant context keys via `internal/delivery` (`delivery.GetContextString`)
- call usecase with `r.Context()`
- return `pkg/response` stdlib helpers (`response.WriteSuccess`, `response.WriteSuccessWithPaging`, `response.WriteHTTPError`, `response.WriteJSON`)

Repository pattern:

- hide `sqlx` parameterized SQL queries behind module repositories
- extract active transaction via `tx.DBTXFromContext(ctx)` (falling back to `*sqlx.DB`)
- apply tenant/visibility clauses via `pkg/database` (`SQLOrganizationClause`, `SQLOrganizationVisibilityClause`) and soft-delete filters (`deleted_at = 0`)
- keep dynamic filtering through `pkg/querybuilder` (`BuildRawQuery`) instead of raw user-controlled SQL fragments

## Worker path

- task distributor: `internal/worker/distributor.go`
- task processor: `internal/worker/processor.go`
- handlers: `internal/worker/handlers/*`
- tasks: `internal/worker/tasks/*`
- scheduler: `internal/worker/scheduler.go`

Worker-triggering modules confirmed from wiring and usecases include auth, audit, organization invite/member flows, and webhook dispatch.

Worker change checklist:

- inspect task type definitions under `internal/worker/tasks`
- inspect distributor method and processor handler registration
- inspect handler side effects and retry/idempotency assumptions
- use integration/E2E tests when request behavior depends on async side effects

## Middleware map

- auth token/session: `internal/middleware/auth_middleware.go`
- tenant organization: `internal/middleware/tenant_middleware.go`
- Casbin authorization: `internal/middleware/casbin_middleware.go`
- API key: `internal/middleware/api_key_middleware.go`
- CORS/security headers/rate limit/recovery/request logging/user status/prometheus: `internal/middleware/*`

Observed middleware layering in `internal/router/http_router.go`:

- `apiKeyAuth` (`HTTPAuthenticate`) runs before `authMw` (`HTTPValidateToken`) on `authChain`, `tenantChain`, `adminChain`, and `apiKeyChain`.
- `csrf` (`HTTPCSRFMiddleware`) and `apiKeyAutoScope` / `HTTPRequireScopes` run after auth and before `userStatus` (`HTTPUserStatusMiddleware`).
- tenant requirement (`HTTPRequireOrganization` / `HTTPOptionalOrganization`) is enforced before `casbinMw` (`HTTPCasbinMiddleware`) on tenant and admin routes.
- upload route uses `authMw` + `userStatus` with base wrappers (`reqID`, `reqLog`, `rec`, `sec`, `cors`), without tenant/Casbin chain middleware.

## Tests

- unit tests colocated under `internal`, `pkg`
- integration tests under `tests/integration`
- E2E tests under `tests/e2e`

Important backend test coverage areas already present:

- auth and middleware
- tenant isolation
- permission and role orchestration
- API key lifecycle
- TUS upload
- webhook dispatch
- realtime SSE/WebSocket
- worker integration

Backend change checklist:

- route/middleware change: inspect `internal/router/http_router.go` and target middleware tests
- module behavior change: inspect target module unit tests and integration tests
- tenant/Casbin change: inspect tenant isolation, permission, role, and transactional enforcer tests
- upload change: inspect `pkg/tus`, upload tests, and storage provider tests
- worker change: inspect `internal/worker` tests and worker integration scenarios
