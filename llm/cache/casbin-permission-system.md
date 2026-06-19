# Casbin Permission System

## Purpose

Durable map for authorization behavior in this repo:

- Casbin boot and production guard
- route-layer authorization order
- permission module ownership
- transactional policy writes
- user, role, access-right, and domain coupling
- batch permission checks and admin routes

Use this file before changing `internal/modules/permission`, `internal/middleware/casbin_middleware.go`, route groups, access-right registry, or any authz-sensitive write flow.

## Primary source of truth

1. `internal/config/app.go`
2. `internal/middleware/casbin_middleware.go`
3. `internal/router/router.go`
4. `internal/modules/permission/*`
5. `internal/modules/permission/usecase/transactional_enforcer.go`
6. `llm/cache/access-right-system.md`
7. `llm/cache/role-system.md`
8. `llm/cache/permission-system.md`

## Runtime ownership

### Boot and fail-open guard

`internal/config/app.go` proves:

- app builds global Casbin enforcer during startup
- in production, startup aborts if enforcer nil
- in production, startup aborts if loaded policy count is zero
- non-production can warn when Casbin disabled, but that is explicit degraded mode

This is repo-critical safety logic. Do not bypass it in docs or code.

### Middleware ownership

`internal/middleware/casbin_middleware.go` owns HTTP request enforcement.

Important behavior:

- release mode logs critical error if enforcer nil
- request object path has trailing slash stripped before enforce call
- enforcement uses subject + domain + object + action
- failed enforce returns forbidden, not silent pass-through

### Permission module ownership

`internal/modules/permission` owns:

- policy CRUD
- role permission grant/revoke
- user role assignment
- access-right expansion inputs
- batch permission check behavior
- cleanup patterns tied to roles and permissions

Do not split effective ownership across controllers or unrelated modules.

## Route-layer authorization model

### `tenantAuthorized`

Current order in `internal/router/router.go`:

1. API-key authenticate
2. token validate
3. API-key auto scope
4. user status
5. require organization
6. Casbin middleware

Implication:

- tenant context is available before Casbin check
- Casbin domain input depends on organization resolution
- API key scope and Casbin can both restrict same request

### `authorized`

Current order:

1. API-key authenticate
2. token validate
3. explicit `admin:manage` API-key scope
4. user status
5. optional organization
6. Casbin middleware

Implication:

- admin routes are not scope-only
- optional org context can still affect domain-aware enforcement

### Authenticated batch check exception

- permission batch-check route is registered under `authenticated`, not only admin routes
- do not assume every permission endpoint is same access tier

## Transactional enforcer behavior

`internal/modules/permission/usecase/transactional_enforcer.go` is key repo-specific behavior.

Proven semantics:

- wrapper holds global enforcer for normal calls
- `WithContext(ctx)` checks transaction-bound DB from `pkg/tx`
- when tx DB exists, wrapper builds transient Casbin enforcer on transaction adapter
- transient enforcer autosaves against transaction-bound adapter

Why this matters:

- DB state and Casbin policy state can commit atomically in some flows
- policy writes inside transaction-sensitive usecases must not accidentally escape transaction boundary

## Enforcement inputs

Effective decision depends on all of these, not one field:

- subject: usually `user_id`
- domain: tenant/org or fallback domain
- object: normalized request path
- action: HTTP method
- loaded policies and grouping policies
- access-right expansion and role assignments upstream

## Cross-system coupling

Casbin changes can silently break:

- role CRUD and hierarchy
- permission assignment flows
- access registry semantics
- tenant middleware ordering
- project, webhook, audit, user, and org route access
- API-key layered authorization behavior

Read with:

- `llm/cache/access-right-system.md`
- `llm/cache/role-system.md`
- `llm/cache/permission-system.md`
- `llm/cache/tenant-organization-system.md`

## Known sharp edges

- route trailing slash normalization happens before enforce; route matcher work must remember this.
- production safety relies on loaded policies, not merely enabled flag.
- optional-org admin routes can still produce domain-sensitive checks.
- permission writes tied to DB state need transactional enforcer, not raw global enforcer shortcuts.

## Change checklist

Before editing Casbin or permission code, prove these answers:

1. Is change on route order, middleware logic, policy data, matcher/model, or permission module usecase?
2. Does target flow require transaction-bound Casbin writes?
3. Does route depend on required org context or optional org context?
4. Could API-key scope already reject or allow before Casbin runs?
5. Will access-right or role changes alter effective policies indirectly?
6. Does production startup guard still prevent fail-open behavior?

## Verification paths

Narrow first:

- `internal/modules/permission/test/*`
- `internal/modules/auth/repository/casbin_adapter_test.go`
- `internal/middleware/casbin_middleware.go`
- `internal/modules/permission/usecase/transactional_enforcer.go`
- `internal/router/router.go`
- `internal/config/app.go`

Broader when route behavior changed:

- package tests for affected module
- integration auth/tenant/Casbin path where domain resolution matters

## Hard rules

- Do not bypass production nil/zero-policy guard.
- Do not move authorization truth into frontend-only checks.
- Do not weaken subject/domain/object/action semantics casually.
- Do not write policy state outside transactional path when DB mutation must stay atomic.
- Do not reorder tenant resolution after Casbin on tenant-scoped routes.
