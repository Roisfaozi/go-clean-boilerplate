---
name: api-endpoint
description: Use when adding or changing backend HTTP endpoints, route protection, API-key scope behavior, Swagger-visible contracts, or frontend-consumed API shapes in this Casbin monorepo.
---

# API Endpoint

## Overview

Endpoint work in this repo is never only controller code.

Every endpoint decision can affect route strata, middleware layering, API-key scope, Swagger output, frontend proxies, and shared types.

## When To Use

Use this skill when:

- adding or changing Gin routes
- moving route between protection strata
- changing request or response contract used by frontend
- changing endpoint scope behavior, tenant semantics, or Swagger-visible API

Use another skill when:

- pure backend usecase logic changes with no route or contract impact; use `go-service`
- transaction-heavy persistence change dominates; use `database-transactions`
- auth, tenant, or Casbin boundary dominates; also load `auth-tenant-casbin`
- API-key decision dominates; also load `api-key-scope`

## Required Read Order

1. `AGENTS.md`
2. `llm/cache/api-contracts.md`
3. `llm/cache/backend-map.md`
4. `llm/cache/domain-rules.md`
5. `llm/workflows/api-endpoint.md`
6. `internal/router/router.go`
7. target `internal/modules/*/delivery/http/*routes.go`
8. target controller, request/response structs, usecase, repository
9. `apps/web/src/app/api/v1/[...path]/route.ts` if `apps/web` can consume endpoint
10. `apps/client/app/routes/api-proxy.ts` if `apps/client` can consume endpoint
11. `packages/api-types/*` when shared contract exists

## Route Strata Matrix

Choose intentionally:

- `public`
  - no auth middleware
  - public auth, invitation, or health-style surfaces only
- `authenticated`
  - API-key authenticate
  - token validation
  - auto scope
  - user session requirement
  - user status middleware
- `tenantAuthorized`
  - authenticated base
  - required organization context
  - Casbin enforcement
- `authorized`
  - admin-style explicit `admin:manage`
  - optional organization context
  - Casbin enforcement
- `upload`
  - upload-specific middleware and TUS path behavior

## Runtime Truth To Preserve

- route strata and middleware layering are defined centrally in `internal/router/router.go`
- some module routes are registered via module route helpers, while some protected groups like project routes are declared directly in router
- frontend consumers may depend on backend endpoints through `apps/web` proxy, `apps/client` proxy, and `packages/api-types`
- Swagger-visible public API changes can require docs regeneration through `pnpm go:docs`

## Endpoint Workflow

### Step 1 — Recon

Find nearest precedent in same module.

Trace route registration from `internal/router/router.go` to module routes or direct route group.

Identify consumers in:

- `apps/web`
- `apps/client`
- `packages/api-types`

### Step 2 — Define Contract

Before patching, write down:

- method and path
- path/query/body params
- response shape and error shape
- route stratum
- API-key scope behavior
- tenant or organization semantics
- whether Swagger artifacts should change

### Step 3 — Patch By Layer

- request parsing and validation in controller
- business rules in usecase
- GORM/query details in repository
- router owns route exposure and protection
- shared type or proxy updates when consumers depend on contract

### Step 4 — Consumer Sync

If contract changes, audit and update:

- `apps/web/src/app/api/v1/[...path]/route.ts`
- `apps/client/app/routes/api-proxy.ts`
- `packages/api-types/*`
- any frontend loader, action, or hook consuming payload

### Step 5 — Verification

Start narrow:

- module controller tests
- usecase tests
- route-specific tests if present

Escalate when needed:

- integration/E2E for route strata, cookies, tenant, API-key, or Casbin behavior
- `pnpm go:docs` when Swagger-visible contract changes

## Review Checklist

- [ ] route registered exactly once
- [ ] route group matches product intent
- [ ] API-key decision is explicit
- [ ] JWT is not treated as enough without session rules when route needs them
- [ ] tenant context comes from middleware/usecase boundary, not ad hoc request fields
- [ ] frontend consumers updated or proven unaffected
- [ ] Swagger artifacts updated or proven unnecessary

## Common Mistakes

- adding handler and forgetting router registration path
- choosing weaker route group for convenience
- changing payload shape without proxy or shared type review
- putting scope or tenant checks in controller body that middleware should own

## Stop Conditions

- stop if route ownership or stratum is unclear
- stop if cache/docs contradict live router layering in a way that changes endpoint design
- stop before destructive DB/schema/data operations not explicitly requested

## Completion Output

Report:

- endpoint contract changed
- route stratum and scope decision
- consumer sync performed
- files changed
- commands run and exact result
- skipped verification and blocker
