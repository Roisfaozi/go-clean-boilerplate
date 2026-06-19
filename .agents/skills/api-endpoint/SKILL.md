---
name: api-endpoint
description: Use when adding or changing backend HTTP endpoints, route protection, API-key scope behavior, Swagger-visible contracts, or frontend-consumed API shape in the Casbin monorepo.
---

# API Endpoint: Route Contract and Protection

**Announce at start:** "I'm using the api-endpoint skill to preserve route, auth, tenant, and contract boundaries."

## When To Use

| Use this skill when...                         | Use another skill when...                            |
| ---------------------------------------------- | ---------------------------------------------------- |
| adding/changing Gin routes                     | pure usecase logic only -> `go-service`              |
| changing route group/auth/Casbin/API-key scope | persistence-only change -> `database-transactions`   |
| changing contract consumed by frontend         | frontend-only UI change -> `frontend-surface` / `ui` |

## Read Order

1. `AGENTS.md`
2. `llm/cache/api-contracts.md`
3. `llm/cache/backend-map.md`
4. `llm/cache/domain-rules.md`
5. `llm/workflows/api-endpoint.md`
6. `internal/router/router.go`
7. target `internal/modules/*/delivery/http/*routes.go`
8. target controller, usecase, repository, request/response structs

## Route Strata Decision

Choose one intentionally:

- `public`: no auth; safe public auth/invitation style flows only.
- `authenticated`: API-key/JWT/session/status checks, no required tenant/Casbin policy.
- `tenantAuthorized`: auth + tenant org + Casbin policy.
- `authorized`: admin-style scope + optional tenant + Casbin policy.
- `upload`: TUS route with upload-specific handler and auth/status middleware.

## Workflow

### Phase 1 — Recon

- Find existing similar endpoint in same module.
- Trace route registration from `internal/router/router.go` to module routes.
- Identify consumers in `apps/web`, `apps/client`, and `packages/api-types`.

### Phase 2 — Contract

Define:

- method and path
- path/query/body params
- response shape and error shape
- route stratum
- API-key scope behavior (auto or explicit)
- tenant/Casbin domain semantics

### Phase 3 — Implementation

- Add/update request/response structs and validation tags.
- Keep parsing/validation in controller.
- Keep business rules in usecase.
- Keep GORM/query details in repository.
- Update Swagger comments/artifacts when public API contract changes.

### Phase 4 — Consumer Sync

- Audit `apps/web/src/app/api/v1/[...path]/route.ts` if `apps/web` can call it.
- Audit `apps/client/app/routes/api-proxy.ts` if `apps/client` can call it.
- Update shared types/client helpers when payload shape changes.

### Phase 5 — Verification

- Narrow module/controller/usecase tests first.
- Integration/E2E when route stratum, cookie/session, tenant, API-key, or Casbin behavior changed.
- `pnpm go:docs` when Swagger artifacts should change.

## Review Checklist

- [ ] route registered exactly once
- [ ] route group matches product intent
- [ ] API-key scope cannot be bypassed
- [ ] JWT parsing is not treated as enough without Redis session validation
- [ ] tenant context comes from middleware/usecase boundary, not ad hoc query params
- [ ] frontend consumers are updated or confirmed unaffected

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
