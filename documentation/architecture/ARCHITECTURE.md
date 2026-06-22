# Architecture Overview

This document is the main architecture guide for future developers. It explains each runtime layer, the scope of that layer, where the code lives, and what kind of logic belongs there.

Live code remains the final authority. Use this document to understand the system before reading or changing code.

## 1. Runtime Entry Points

| Concern | File | Scope |
| --- | --- | --- |
| API process startup | `cmd/api/main.go` | load config, create application, start/stop HTTP server and background services |
| dependency composition | `internal/config/app.go` | construct DB, Redis, JWT, Casbin, storage, realtime managers, modules, middleware, router, worker, scheduler |
| route and middleware topology | `internal/router/router.go` | define `/api/v1`, public/authenticated/tenant/admin route groups, upload, metrics, realtime endpoints |
| environment config mapping | `internal/config/config.go` | map env/defaults into `AppConfig`; define config structs and defaults |

Developer rule: when behavior seems unclear, read files in this order: `cmd/api/main.go`, `internal/config/app.go`, `internal/router/router.go`, then the target module.

## 2. High-Level System Shape

The project is a polyglot monorepo with a Go API as the operational core and active frontend apps under `apps/*`.

```txt
clients
  ├─ apps/web      Next.js frontend and API proxy
  ├─ apps/client   React Router frontend and API proxy
  └─ external API clients
       │
       ▼
Gin HTTP API
       │
       ├─ middleware layer       auth, API key, tenant, Casbin, rate limit, observability
       ├─ delivery layer         HTTP controllers
       ├─ usecase layer          business orchestration
       ├─ repository layer       GORM / Redis persistence
       ├─ infrastructure layer   JWT, storage, TUS, SSE, WS, SSO, Asynq, telemetry
       └─ worker layer           async cleanup, audit sync, webhook dispatch, scheduled jobs
```

The backend follows Clean Architecture boundaries, but not in a dogmatic way. Modules own their domain code, while `internal/config/app.go` composes cross-module dependencies explicitly.

## 3. Layer Map And Scope

### 3.1 Process / Bootstrap Layer

**Files:**

- `cmd/api/main.go`
- `internal/config/config.go`
- `internal/config/app.go`

**Scope:**

- load environment config
- initialize logger, validator, DB, Redis, telemetry
- build shared managers and providers
- wire modules and middleware
- create HTTP server
- start worker processor and scheduler
- handle graceful shutdown

**Belongs here:**

- dependency construction
- long-lived infrastructure clients
- application lifecycle
- strict startup guards such as Casbin policy guard

**Does not belong here:**

- business rules
- per-request decision logic
- route-specific validation
- repository queries

### 3.2 Router / Access Segmentation Layer

**File:** `internal/router/router.go`

**Scope:**

- global Gin middleware
- route grouping
- route registration
- access-level separation
- upload route wiring
- metrics endpoint wiring

**Access strata:**

| Group | Middleware | Scope |
| --- | --- | --- |
| public | optional public rate limit | login, register, refresh, password reset, email verification, SSO, public invite accept paths |
| authenticated | API key auth, JWT/session validation, automatic API-key scope, user-session requirement, user status check | user self-service, logout, me, tickets, stats, organization list, batch permission check |
| tenant authorized | API key auth, JWT/session validation, scope check, user status, required organization, Casbin | organization tenant routes, projects, tenant webhooks |
| admin authorized | API key auth, JWT/session validation, `admin:manage`, user status, optional organization, Casbin | admin organization operations, permissions, access rights, roles, users, audit |
| upload | JWT/session validation, user status, TUS handler | resumable upload endpoint under `/api/v1/upload/files/*any` |

**Belongs here:**

- middleware order
- route group ownership
- explicit scope gate at route boundary
- registration of module HTTP routes

**Does not belong here:**

- data filtering rules that belong in repositories/usecases
- UI-only permission assumptions
- tenant membership logic internals

### 3.3 Middleware Layer

**Files:** `internal/middleware/*`

**Scope:**

- request cross-cutting concerns
- authentication and API-key identity
- tenant context resolution
- Casbin authorization
- user status gate
- request ID, logging, recovery, security headers, CORS
- rate limiting, metrics, OTEL

**Key boundaries:**

- `AuthMiddleware` validates backend-authenticated user/session state and realtime tickets.
- `APIKeyMiddleware` authenticates API-key callers and enforces API-key scopes.
- `TenantMiddleware` resolves organization context and verifies membership.
- `CasbinMiddleware` enforces resource/action/domain policy after identity and tenant context exist.
- `UserStatusMiddleware` blocks disabled, banned, or invalid user states before domain handlers run.

**Belongs here:**

- checks that must happen before handlers
- context injection for downstream code
- request-level security decisions

**Does not belong here:**

- persistence mutation
- business workflows
- frontend route guarding as a substitute for backend enforcement

### 3.4 Delivery / Controller Layer

**Files:** `internal/modules/*/delivery/http/*`

**Scope:**

- parse request path/query/body
- call validator
- read middleware-injected context
- call usecase interfaces
- shape HTTP responses
- map domain errors to HTTP status codes

**Belongs here:**

- request/response DTO mapping
- HTTP-specific behavior
- binding and validation entry
- selecting usecase method

**Does not belong here:**

- transaction orchestration unless the usecase explicitly exposes it
- raw DB access
- cross-module business rules
- Casbin policy mutation logic

### 3.5 Usecase Layer

**Files:** `internal/modules/*/usecase/*`

**Scope:**

- application business rules
- domain orchestration
- transaction boundaries
- side-effect sequencing
- calls across repositories, storage, audit, webhook, worker, and Casbin interfaces

**Belongs here:**

- create/update/delete business rules
- auth workflows such as login/register/reset/SSO acceptance
- tenant membership and organization lifecycle decisions
- permission assignment and role orchestration
- upload completion hooks such as avatar update
- audit/webhook enqueue decisions

**Does not belong here:**

- HTTP binding details
- SQL string construction that belongs in repositories
- frontend state decisions
- raw environment config struct passing

Important repo rule: usecase constructors should receive only needed dependencies/values, not the full `AppConfig`.

### 3.6 Repository Layer

**Files:** `internal/modules/*/repository/*`

**Scope:**

- GORM persistence
- Redis persistence/cache where domain-owned
- query construction
- tenant-scoped filtering
- field allowlists for dynamic query behavior
- persistence model/entity conversion

**Belongs here:**

- DB queries
- transaction-aware repository methods
- organization visibility scopes
- soft-delete filters
- querybuilder allow/deny rules

**Does not belong here:**

- HTTP context parsing
- permission decisions unrelated to persistence
- business workflows that coordinate multiple repositories

### 3.7 Entity / Model Layer

**Files:**

- `internal/modules/*/entity/*`
- `internal/modules/*/model/*`

**Scope:**

- domain entities
- database models
- request/response models
- DTOs used by module boundaries

**Belongs here:**

- stable shapes for module contracts
- persistence fields and relation tags where applicable
- explicit conversion helpers if module uses them

**Does not belong here:**

- runtime service dependencies
- database query execution
- HTTP handler logic

### 3.8 Infrastructure / Shared Package Layer

**Files:** `pkg/*`

**Scope:**

Shared low-level building blocks used by modules and app wiring.

| Package | Scope |
| --- | --- |
| `pkg/jwt` | access/refresh token creation and parsing |
| `pkg/tx` | transaction manager abstraction |
| `pkg/storage` | local and S3-compatible storage providers |
| `pkg/tus` | resumable upload server, metadata validation, upload hook dispatch |
| `pkg/ws` | WebSocket manager, tickets, Redis presence, distributed realtime |
| `pkg/sse` | SSE manager and broadcast handling |
| `pkg/querybuilder` | safe dynamic filters/sorts/search field handling |
| `pkg/telemetry` | OTEL tracer initialization |
| `pkg/sso` | SSO provider integrations |
| `pkg/circuitbreaker` | global circuit breaker setup |

**Belongs here:**

- reusable infrastructure behavior
- code that is not domain-specific to one module
- helper abstractions with clear runtime ownership

**Does not belong here:**

- module-specific business workflows
- frontend-specific code
- route-specific authorization policy

### 3.9 Worker / Async Layer

**Files:**

- `internal/worker/*`
- `internal/worker/handlers/*`
- `internal/worker/tasks/*`

**Scope:**

- Asynq task distribution
- task processing
- scheduled cleanup jobs
- webhook delivery
- audit sync
- email-like background operations

**Belongs here:**

- async retryable side effects
- scheduled maintenance
- worker handlers that consume task payloads

**Does not belong here:**

- synchronous request authorization
- logic that must complete before HTTP response unless usecase waits for it

Developer rule: HTTP success that enqueues a worker task means request accepted the side effect, not necessarily that the async side effect has completed.

### 3.10 Frontend / API Consumer Layer

**Files:**

- `apps/web`
- `apps/client`
- `packages/*`

**Scope:**

- user-facing UI
- frontend route handling
- API proxies
- typed frontend helpers
- shared UI/components/types/utilities

**Belongs here:**

- UI composition
- client/server API calls
- frontend auth UX
- active organization selection
- user-visible loading/error/empty states

**Does not belong here:**

- security enforcement as the only gate
- backend tenant authorization
- Casbin policy decisions

Backend remains source of truth for auth, tenant, API-key, and permission enforcement.

## 4. Module Ownership

| Module | Primary Scope | Notes |
| --- | --- | --- |
| `auth` | registration, login, refresh, reset, SSO, session/ticket behavior | coordinates JWT, Redis sessions, audit, realtime, org bootstrap, Casbin default role/domain |
| `user` | users, profile, avatar, status | integrates storage/TUS hooks, audit, webhook, auth side effects |
| `organization` | tenants, members, invitations, lifecycle | owns org membership reads and tenant lifecycle behavior |
| `permission` | permission policy CRUD, batch checks, Casbin policy writes | uses transactional enforcer where DB and policy must commit together |
| `role` | role CRUD and role-permission orchestration | depends on permission usecase for policy changes |
| `access` | access-right registry and endpoint grouping | feeds permission expansion and API-key scope mapping |
| `project` | tenant-scoped projects | must remain organization-scoped |
| `api_key` | organization-aware API-key auth and scope checks | used at middleware and route levels |
| `audit` | audit persistence, stream/broadcast, audit worker sync | request side effects and worker coupling are high-risk |
| `webhook` | webhook CRUD and async delivery | dispatch is worker-backed |
| `stats` | dashboard summary/activity/insights | also feeds realtime metrics broadcaster |

## 5. Cross-Cutting Security Boundaries

### Authentication

Authentication is not just JWT parsing. The backend validates active session state and user status before sensitive handlers run.

### API Keys

API-key requests pass through API-key middleware before route handlers. Scope checks can be automatic or explicit route-level checks such as `project:manage`.

### Tenant Context

Tenant routes require organization context. The tenant layer resolves headers such as organization ID/slug, verifies membership, and injects organization context for downstream code.

### Casbin Authorization

Casbin runs after identity and tenant context are available. In strict environments, startup fails if Casbin is disabled or policies are missing.

### Query Safety

Dynamic list/search endpoints must use field allowlists and avoid exposing sensitive fields through querybuilder.

### Upload Safety

TUS upload metadata controls hook behavior. Upload hooks must validate ownership and context before mutating domain state such as avatar URLs.

### Async Side Effects

Audit, webhook, cleanup, and email-like effects can run through workers. Treat enqueue and completion as separate states.

## 6. Request Flow Summary

A typical protected tenant request follows this order:

1. Gin global middleware: recovery, request ID, metrics, logger, security headers, CORS, optional rate limit.
2. API-key middleware: authenticate API key if present.
3. Auth middleware: validate user token/session.
4. Scope middleware: auto or explicit API-key scope enforcement.
5. User status middleware: block inactive/banned/suspended users.
6. Tenant middleware: require organization and verify membership.
7. Casbin middleware: enforce domain-aware permission.
8. Controller: parse request, validate input, call usecase.
9. Usecase: execute business rule, open transaction if needed, call repositories/infrastructure.
10. Repository: persist/read data through GORM or Redis.
11. Side effects: audit, webhook, realtime broadcast, worker enqueue, storage/TUS hook as needed.
12. Controller: return HTTP response.

## 7. Transaction And Policy Rules

- Database writes that must commit with Casbin policy changes should use transaction-aware enforcer behavior.
- Storage writes should receive `context.Context` and respect cancellation/deadline.
- Worker enqueues should happen only after the usecase has enough durable state to retry safely.
- If a module performs multiple writes, use `pkg/tx` or explicit transaction-aware repository methods.

## 8. Documentation Map

- `SYSTEM_ARCHITECTURE.md` — shorter compatibility summary.
- `MULTI_TENANCY.md` — tenant, membership, GORM scope, Casbin domain architecture.
- `ARCHITECTURE_VISUAL_AND_SEQUENCE.md` — visual diagrams and core request sequences.
- `../guides/DEVELOPER_FLOW.md` — reading and verification workflow.
- `../guides/API_ACCESS_WORKFLOW.md` — API auth/RBAC workflow.
- `../guides/TESTING.md` — test strategy.
