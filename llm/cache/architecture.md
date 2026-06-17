# Architecture

## Source of truth order

1. `cmd/api/main.go`
2. `internal/config/app.go`
3. `internal/router/router.go`
4. `internal/modules/*/module.go`
5. module delivery/usecase/repository files
6. tests under `internal`, `pkg`, and `tests`
7. documentation under `documentation/`

## Runtime composition

`cmd/api/main.go` starts the API binary and calls application construction through `internal/config.NewApplication`.

`internal/config/app.go` is the composition root. It wires:

- logger
- circuit breaker config
- OpenTelemetry tracer
- validator
- GORM database connection
- Redis client
- Asynq task distributor and processor
- transaction manager
- JWT manager
- WebSocket presence and ticket managers
- WebSocket manager
- SSE manager
- Casbin enforcer
- storage provider
- TUS upload handler and hook registry
- SSO providers
- business modules
- Gin router
- scheduler

The composition root also initializes production safety around Casbin and shared managers for realtime, auth, and upload flows before route registration.

## HTTP boundary

`internal/router/router.go` creates the Gin router and route groups.

Top-level surfaces:

- `/api/v1/docs/*any`: Swagger UI
- `/metrics`: Prometheus metrics, optionally basic-auth protected
- `/api/v1/events`: SSE, protected by auth token
- `/api/v1/ws`: WebSocket, protected by WebSocket ticket validation
- `/api/v1/health`: health check
- `/api/v1/auth/*`: public and authenticated auth routes
- tenant and Casbin-protected module routes under `/api/v1`
- `/api/v1/upload/files/*`: TUS upload endpoint mounted outside normal REST controller routing

## Route authorization strata

`internal/router/router.go` defines practical security strata:

1. public group: public auth, public user/org invitation flows
2. authenticated group: API key middleware, JWT validation, scope auto-check, user session requirement, user status middleware
3. tenantAuthorized group: authenticated + organization requirement + Casbin middleware
4. admin/authorized management groups: route-specific Casbin/API-key scope checks for permission, role, access, audit, webhook, and tenant resources

Operational route details:

- `public` group handles unauthenticated auth and public organization/user flows.
- `authenticated` group adds API-key auth, JWT/session validation, scope checks, and user-status enforcement.
- `tenantAuthorized` adds organization resolution plus Casbin policy checks.
- `authorized` is for admin-style management with `admin:manage` scope and optional organization context.
- TUS upload group uses token auth + user status only, then delegates to `pkg/tus`.

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

- GORM is the runtime ORM.
- Migration files live under `db/migrations`.
- Seed entrypoint lives under `db/seeds/main.go`.
- The repo uses MySQL driver in `go.mod`.

## Frontend boundary

Two active frontend applications exist:

- `apps/web`: Next.js App Router, API proxy routes under `apps/web/src/app/api`.
- `apps/client`: React Router, route config in `apps/client/app/routes.ts`, API proxy route `api/v1/*`.

Frontend integration pattern:

- `apps/web` owns server-side proxy and auth callback/logout route handlers.
- `apps/client` owns route registry and client-side API proxy route.
- both apps rely on shared workspace packages for reusable components, hooks, API types, and utilities.

Frontend architecture details are documented in `llm/cache/frontend-map.md`.
