# Backend Map

## Entrypoints

- API binary: `cmd/api/main.go`
- Generator binary: `cmd/gen/main.go`
- Composition root: `internal/config/app.go`
- Router: `internal/router/router.go`

## Shared infrastructure

- config loading: `internal/config/config.go`
- DB: `internal/config/gorm.go`
- Redis: `internal/config/redis.go`
- Casbin: `internal/config/casbin.go`, `internal/config/casbin_model.conf`
- storage: `internal/config/storage.go`, `pkg/storage/*`
- transaction manager: `pkg/tx`
- JWT: `pkg/jwt`
- SSE: `pkg/sse`
- WebSocket: `pkg/ws`
- TUS: `pkg/tus`
- telemetry: `pkg/telemetry`
- query builder: `pkg/querybuilder`

Runtime wiring direction:

- `internal/config/app.go` owns cross-module composition.
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

1. Gin route in `internal/router/router.go` or module route file.
2. Middleware applies auth, API-key, tenant, status, rate-limit, and/or Casbin checks.
3. Controller parses and validates request.
4. Usecase executes business behavior and transaction handling.
5. Repository reads/writes GORM models.
6. Side effects go through worker, audit, webhook, SSE, or WebSocket managers.

This pattern is visible in module constructors and route/controller files, not only in documentation.

Controller pattern:

- bind request using Gin (`ShouldBindJSON`, path/query params, or upload-specific handling)
- validate with injected validator where applicable
- call usecase with `context.Context` from request
- return project response helpers / JSON responses

Repository pattern:

- hide GORM query details behind module repositories
- use scopes/context for tenant-sensitive data where module design requires it
- keep dynamic filtering through `pkg/querybuilder` instead of raw user-controlled SQL fragments

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

Observed middleware layering in router:

- API key authenticate happens before JWT validation on authenticated, tenantAuthorized, and authorized groups.
- user status middleware is applied after auth/API-key identity resolution.
- tenant requirement is enforced before Casbin on tenant-authorized routes.
- upload route uses auth + user-status middleware but not tenant/Casbin group middleware.

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

- route/middleware change: inspect `internal/router/router.go` and target middleware tests
- module behavior change: inspect target module unit tests and integration tests
- tenant/Casbin change: inspect tenant isolation, permission, role, and transactional enforcer tests
- upload change: inspect `pkg/tus`, upload tests, and storage provider tests
- worker change: inspect `internal/worker` tests and worker integration scenarios
