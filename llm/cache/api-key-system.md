# API Key System

## Purpose

Durable map for API-key behavior in this repo:

- API-key creation, listing, revoke
- request authentication from `X-API-Key`
- scope derivation and enforcement
- organization-scoped identity semantics
- interaction with JWT auth, user session rules, tenant rules, and Casbin-protected routes

Use this file before changing `internal/modules/api_key`, route protection, scope names, tenant route access, or frontend/admin flows that manage API keys.

## Primary source of truth

1. `internal/router/router.go`
2. `internal/middleware/api_key_middleware.go`
3. `internal/modules/api_key/usecase/api_key_usecase.go`
4. `internal/modules/api_key/delivery/http/api_key_routes.go`
5. `internal/modules/api_key/repository/*`
6. `llm/cache/project-system.md` for project route coupling
7. `llm/cache/authentication-system.md` and `llm/cache/tenant-organization-system.md`

## Runtime ownership

### Route ownership

- API-key management routes are registered by `api_keyHttp.RegisterApiKeyRoutes(authenticated, ...)`.
- this means API-key CRUD itself is not public and not pure machine-auth; caller still enters authenticated route group.
- project routes under `tenantAuthorized` add explicit API-key scopes like `project:view` and `project:manage` on top of auto scope middleware.
- admin route group `authorized` requires explicit `admin:manage` API-key scope before optional org context and Casbin middleware.

### Middleware layering

`internal/router/router.go` proves layering order matters:

#### `authenticated`

1. `apiKeyMiddleware.Authenticate()`
2. `authMiddleware.ValidateToken()`
3. `apiKeyMiddleware.RequireScopeAuto()`
4. `apiKeyMiddleware.RequireUserSession()`
5. `UserStatusMiddleware`

Implication:

- API key may identify caller first.
- JWT/session validation still runs on authenticated group.
- pure API-key access is blocked on routes that require user session.

#### `tenantAuthorized`

1. `Authenticate()`
2. `ValidateToken()`
3. `RequireScopeAuto()`
4. `UserStatusMiddleware`
5. `tenantMiddleware.RequireOrganization()`
6. `casbinMiddleware`

Implication:

- API key alone does not bypass tenant membership or Casbin.
- organization context still matters.

#### `authorized`

1. `Authenticate()`
2. `ValidateToken()`
3. `RequireScopes("admin:manage")`
4. `UserStatusMiddleware`
5. `tenantMiddleware.OptionalOrganization()`
6. `casbinMiddleware`

Implication:

- admin-style endpoints still stack API-key scope and Casbin, not either/or.

## Identity injection behavior

`internal/middleware/api_key_middleware.go` proves successful API-key auth injects:

- `user_id`
- `organization_id`
- `username`
- `auth_method = api_key`
- `api_key_id`
- `api_key_scopes`

If `organization_id` exists, middleware also writes org context into request context via `pkg/database.SetOrganizationContext`.

This means API-key auth changes can affect downstream repository scoping even when handler code never reads API-key fields directly.

## Scope enforcement model

### Auto scope

`RequireScopeAuto()` derives scope from request path and method:

- resource taken from `/api/v1/:resource`
- trailing `s` singularized with basic rule
- method maps through `ScopeFromMethod`

Examples from code behavior:

- `GET /api/v1/projects` -> `project:view`
- `POST /api/v1/users` -> `user:create`

Auto scope skips only when:

- request is not API-key-auth
- path has fewer than 3 trimmed parts

### Explicit scope

`RequireScopes()` allows OR-style matching across provided scopes.

Project routes prove repo uses explicit scopes to avoid relying only on coarse auto path logic.

### All-scope mode

`RequireAllScopes()` exists for AND-style checks on sensitive endpoints.

Audit any future use carefully; current route wiring should prove when AND semantics are really needed.

### Wildcards

Tests in `internal/middleware/api_key_middleware_test.go` prove wildcard behavior exists:

- `project:*` can satisfy `project:manage`
- `*` can satisfy arbitrary scopes

Treat wildcard changes as security-critical.

## API key lifecycle behavior

`internal/modules/api_key/usecase/api_key_usecase.go` owns:

- key creation
- scope JSON serialization and parsing
- revoke behavior
- authenticate/identity resolution
- organization status checks and related cache lookups

Important implication:

- scope persistence format is repo contract, not only transport detail
- revocation and org-status checks belong in usecase boundary, not middleware-only logic

## Cross-system coupling

API-key changes can break:

- `project` module routes and org-scoped CRUD access
- auth/session semantics on authenticated endpoints
- tenant resolution because org context is injected from key identity
- Casbin route behavior because API key and Casbin can both gate same request
- frontend API-key management screens through backend contract changes

Read with:

- `llm/cache/project-system.md`
- `llm/cache/casbin-permission-system.md`
- `llm/cache/tenant-organization-system.md`

## Known sharp edges

- `Authenticate()` is additive; if header absent it falls through to JWT path.
- `RequireUserSession()` forbids API-key-authenticated access on routes that should remain human-session only.
- path-based auto scope uses basic singularization; non-standard resource naming needs route-specific review.
- API-key identity injects `user_id`, so downstream code may believe request is user-authenticated unless it also checks auth method semantics.

## Change checklist

Before editing API-key code, prove these answers:

1. Is target route human-session only, machine-capable, or mixed?
2. Should route use auto scope, explicit OR scope, or explicit ALL-scope semantics?
3. Does route also require tenant context or Casbin?
4. Does API-key auth need to set or preserve request org context?
5. Could wildcard scope accidentally widen access?
6. Does frontend/admin UI depend on current API-key response shape?

## Verification paths

Narrow first:

- `internal/middleware/api_key_middleware_test.go`
- `internal/modules/api_key/test/*`
- `internal/router/router.go`
- `internal/modules/api_key/usecase/api_key_usecase.go`

Broader when route layering changed:

- target module tests
- integration tests covering auth + tenant + API-key + Casbin path

## Hard rules

- Do not treat API key as replacement for tenant or Casbin checks.
- Do not bypass `RequireUserSession()` on human-session routes without explicit design change.
- Do not rename or remap scopes casually; scope strings are runtime security contract.
- Do not assume path-derived scope is correct for custom resource shapes; verify route intent.
- Do not widen wildcard semantics without security review.
