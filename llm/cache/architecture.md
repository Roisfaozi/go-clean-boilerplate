# Architecture

## Source of truth order

1. `cmd/api/main.go`
2. `internal/config/fx.go` and `internal/config/app.go`
3. `internal/router/http_router.go`
4. `internal/modules/*/module.go`
5. module delivery/usecase/repository files
6. tests under `internal`, `pkg`, and `tests`
7. documentation under `documentation/`

## Runtime composition

`cmd/api/main.go` validates startup config and boots `config.CoreFxModule` via `go.uber.org/fx` (`internal/config/fx.go`), which calls `internal/config.NewApplication` and registers `RegisterAppLifecycle` startup/shutdown hooks.

`internal/config/app.go` is the top-level composition root. Most wiring helpers now live in `internal/config/app_helpers.go`, but `app.go` still owns the final assembly. It wires:

- logger
- circuit breaker config
- OpenTelemetry tracer
- validator
- `sqlx` database connection (`*sqlx.DB` via `NewSQLXDatabase` in `internal/config/sqlx.go` with `otelsql` instrumentation)
- Redis client
- Asynq task distributor and processor
- `sqlx` transaction manager (`tx.NewSQLXTransactionManager`)
- JWT manager
- WebSocket presence and ticket managers
- WebSocket manager
- SSE manager
- Casbin enforcer (`pkg/casbinadapter.NewSQLXCasbinAdapter`)
- storage provider
- TUS upload handler and hook registry
- SSO providers
- business modules
- stdlib `http.Handler` router (`router.SetupStdRouter` in `internal/router/http_router.go`) wrapped in `*http.Server`
- scheduler

`internal/config/app_helpers.go` holds extracted builders for SSO providers, module graphs, realtime metrics broadcasting, and TUS handler setup.

The composition root also initializes production safety around Casbin and shared managers for realtime, auth, and upload flows before route registration.

## HTTP boundary

`internal/router/http_router.go` centrally registers all routes on a Go `http.ServeMux` (`SetupStdRouter`) using `delivery.Chain` middleware chains and stdlib `HTTP*` controller handlers (wrapping with `otelhttp.NewHandler` when OTEL is enabled, and mounting Swagger UI via `internal/router/swagger_handler.go`). Per-module route registration files are no longer used.

Top-level surfaces:

- `/metrics`: Prometheus metrics, optionally basic-auth protected
- `/api/v1/events`: SSE, protected by auth token
- `/api/v1/ws`: WebSocket, protected by WebSocket ticket validation
- `/api/v1/health`: health check
- `/api/v1/auth/*`: public and authenticated auth routes
- tenant and Casbin-protected module routes under `/api/v1`
- `/api/v1/upload/files/*`: TUS upload endpoint mounted outside normal REST controller routing

## Route authorization strata

`internal/router/http_router.go` defines practical security chains:

1. `publicChain`: request ID, logging, recovery, security headers, CORS (public auth, public user/org invitation flows)
2. `authChain`: `publicChain` + API key auth, JWT validation, CSRF check, auto-scope check, user status middleware
3. `tenantChain`: `authChain` + organization requirement + Casbin middleware
4. `adminChain`: `publicChain` + API key auth, JWT validation, CSRF check, `admin:manage` scope check, user status, optional organization context, Casbin middleware
5. `apiKeyChain`: `authChain` + `HTTPRequireUserSession` (rejects API keys) + organization requirement

Operational route details:

- `publicChain` handles unauthenticated auth and public organization/user flows.
- `authChain` adds API-key auth, JWT/session validation, CSRF, scope checks, and user-status enforcement.
- `tenantChain` adds organization resolution plus Casbin policy checks.
- `adminChain` is for admin-style management with `admin:manage` scope and optional organization context.
- TUS upload route uses `HTTPValidateToken` + `HTTPUserStatusMiddleware` with base wrappers, then delegates to `pkg/tus`.

## Module architecture

Modules live under `internal/modules/*` and typically use this flow:

`delivery/http` -> `usecase` -> `repository` -> `entity/model`

The module constructors in `internal/modules/*/module.go` wire repositories, usecases, controllers, validators, logging, and cross-module dependencies.

Constructor quality pattern:

- dependencies are passed explicitly, not via giant global config objects
- modules compose other module repos/usecases only when boundary requires it
- cross-module dependencies are introduced in constructors, not in controllers

Examples of wired dependencies:

- `auth` consumes JWT, token repo, user repo, organization repo, transaction manager, notification publisher, authz adapter, task distributor, ticket manager, and SSO providers.
- `organization` consumes Redis-backed org reader, member/invitation repos, transaction manager, enforcer, task distributor, and presence reader.
- `user` consumes audit, auth, webhook, storage, transaction manager, and enforcer.
- `api_key` consumes org repo, user repo, Redis, and API key repo.

## Async and realtime boundaries

- Workers use Asynq and Redis, wired in `internal/config/app.go` and implemented under `internal/worker`.
- SSE manager is created in `internal/config/app.go` and exposed through `/api/v1/events`.
- WebSocket manager, ticket manager, and presence manager are created in `internal/config/app.go`; the route is `/api/v1/ws`.

Health and observability:

- `/api/v1/health` pings MySQL and Redis and downgrades to `DEGRADED` if one fails.
- `/metrics` exists as a separate observability surface and can be basic-auth protected.

## Storage and upload boundary

- Storage provider is selected in `internal/config/app.go` / storage config.
- TUS handler lives under `pkg/tus`.
- Upload completion hooks are registered from app wiring, e.g. avatar upload integrates with user usecase.

The upload boundary is intentionally separate from standard CRUD routes so upload lifecycle, metadata, and hook dispatch stay isolated from ordinary JSON handlers.

## Database boundary

- `sqlx` (`*sqlx.DB` initialized in `internal/config/sqlx.go` with `otelsql` tracing) with parameterized raw SQL is the runtime database layer.
- `pkg/tx` provides `SQLXTransactionManager`, `DBTX`, and `DBTXFromContext` (`DBFromContext` delegates to `DBTXFromContext`).
- `pkg/database` provides tenant/visibility SQL clause helpers (`SQLOrganizationClause`, `SQLOrganizationVisibilityClause`).
- `pkg/casbinadapter` provides `SQLXCasbinAdapter` for Casbin policy persistence.
- Migration files live under `db/migrations`.
- Seed entrypoint lives under `db/seeds/main.go`.
- The repo uses `github.com/go-sql-driver/mysql` in `go.mod`.

## Frontend boundary

Two active frontend applications exist:

- `apps/web`: Next.js App Router, API proxy routes under `apps/web/src/app/api`.
- `apps/client`: React Router, route config in `apps/client/app/routes.ts`, API proxy route `api/v1/*`.

Frontend integration pattern:

- `apps/web` owns server-side proxy and auth callback/logout route handlers.
- `apps/client` owns route registry and client-side API proxy route.
- both apps rely on shared workspace packages for reusable components, hooks, API types, and utilities.

Frontend architecture details are documented in `llm/cache/frontend-map.md`.
