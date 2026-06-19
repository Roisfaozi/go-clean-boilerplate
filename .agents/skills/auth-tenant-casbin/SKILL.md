---
name: auth-tenant-casbin
description: Use when changing authentication, Redis-backed sessions, organization tenant resolution, membership cache, Casbin policy/enforcement, protected route middleware, or role/permission behavior.
---

# Auth Tenant Casbin: High-Risk Boundary

**Announce at start:** "I'm using the auth-tenant-casbin skill because this touches the repo's highest-risk access boundary."

## Iron Rule

Do not assume JWT validity is enough. Protected routes need the repo's middleware/session/tenant/Casbin layering as implemented in live code.

## Read Order

1. `AGENTS.md`
2. `llm/cache/domain-rules.md`
3. `llm/cache/backend-map.md`
4. `internal/router/router.go`
5. `internal/middleware/auth_middleware.go`
6. `internal/middleware/tenant_middleware.go`
7. `internal/middleware/casbin_middleware.go`
8. target auth/organization/permission/role usecases

## Boundary Checklist

- bearer/cookie token parsing
- Redis-backed session validation
- user status middleware
- organization tenant context
- membership/cache invalidation
- Casbin subject/domain/object/method enforcement
- API-key scope layering if route is protected

## Workflow

### Phase 1 — Trace Request

Trace exact request path:
`router group -> API-key middleware -> auth/session -> user status -> tenant -> Casbin -> controller -> usecase`.

### Phase 2 — Determine Authority

- Auth/session belongs in middleware/usecase boundaries.
- Tenant org belongs in tenant middleware and organization usecase/cache paths.
- Policy writes belong in permission/Casbin abstractions.
- Route protection belongs in router/middleware, not frontend UI.

### Phase 3 — Patch

- Preserve Redis session checks.
- Preserve tenant context requirements before Casbin on tenant routes.
- Preserve owner/admin/member constraints in organization flows.
- Use transactional enforcer patterns for policy writes tied to DB transactions.

### Phase 4 — Verify

- Narrow middleware/usecase tests first.
- Integration/E2E for route lifecycle, cookies/tokens, tenant isolation, role/permission decisions.

## Red Flags

- "JWT parsed successfully" used as full auth proof.
- org ID accepted from request without membership/tenant boundary.
- Casbin policy changed outside permission abstractions.
- route moved to public/authenticated group for convenience.

## Stop Conditions

- Stop and ask before destructive DB/schema/data operations not explicitly requested.
- Stop if live code contradicts `llm/cache/*`; live code wins, then document drift in `llm/tasks/`.
- Stop if route ownership, tenant boundary, or auth stratum is unclear.

## Completion Output

Report:

- files changed
- commands run and exact result
- verification skipped and exact blocker
- risks or follow-up work
