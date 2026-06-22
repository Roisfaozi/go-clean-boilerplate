---
name: auth-tenant-casbin
description: Use when changing authentication, Redis-backed sessions, organization tenant resolution, membership cache, Casbin enforcement or policy behavior, or protected-route middleware layering in this Casbin repo.
---

# Auth Tenant Casbin

## Overview

This is highest-risk boundary skill in repo.

These changes can silently weaken access control if router group, middleware order, session checks, tenant resolution, or Casbin enforcement drift even slightly.

## Iron Rule

Do not treat parsed JWT as full authentication proof.

Protected behavior in this repo depends on live middleware layering, Redis-backed session checks, user status checks, tenant resolution, and sometimes Casbin plus API-key scope.

## When To Use

Use this skill when:

- changing auth middleware or auth usecase behavior
- changing session validation, ticket, logout, refresh, or SSO behavior
- changing tenant resolution or organization membership checks
- changing Casbin middleware, permission enforcement, or policy-related protected flow
- moving routes between `public`, `authenticated`, `tenantAuthorized`, or `authorized`

## Read Order

1. `AGENTS.md`
2. `llm/cache/domain-rules.md`
3. `llm/cache/authentication-system.md`
4. `llm/cache/tenant-organization-system.md`
5. `llm/cache/casbin-permission-system.md`
6. `llm/cache/role-system.md`
7. `internal/router/router.go`
8. `internal/middleware/auth_middleware.go`
9. `internal/middleware/tenant_middleware.go`
10. `internal/middleware/casbin_middleware.go`
11. `internal/middleware/api_key_middleware.go` when protected route also accepts API-key actor
12. target auth, organization, permission, or role usecase paths

## Runtime Truth To Preserve

- `authenticated` routes use API-key authenticate, token validation, auto scope, user-session requirement, and user-status middleware
- `tenantAuthorized` adds required organization context and Casbin enforcement
- `authorized` uses explicit `admin:manage`, optional organization context, and Casbin enforcement
- route protection belongs in router plus middleware layering, not frontend checks and not ad hoc handler logic

## Boundary Checklist

Always classify whether current path depends on:

- bearer token or cookie parsing
- Redis session index / active session validation
- user status middleware
- required vs optional organization context
- membership cache invalidation or refresh
- Casbin subject, domain, object, method enforcement
- API-key identity and scope layering

## Workflow

### Step 1 — Trace Exact Request Path

Trace:

`router group -> API-key middleware -> auth/session -> user status -> tenant -> Casbin -> controller -> usecase`

If actual path differs from expected design, follow live code.

### Step 2 — Classify Authority Owner

Decide where rule belongs:

- token and session validity: auth middleware/usecase
- organization resolution: tenant middleware plus organization/member paths
- authorization decision: Casbin and permission abstractions
- route exposure: router group and middleware

### Step 3 — Patch Without Weakening Boundary

- preserve Redis session checks
- preserve tenant context before Casbin on tenant-scoped routes
- preserve owner/admin/member distinctions in organization-sensitive flows
- preserve explicit admin scope on admin-style routes
- use permission abstractions or transactional enforcer patterns for policy writes tied to DB state

### Step 4 — Verify Narrow Then Broad

Start with closest middleware/usecase tests.

Escalate to integration or E2E when change affects:

- cookie or token lifecycle
- tenant isolation
- route group movement
- role or permission decision outcomes

## Common Mistakes

- using successful JWT parse as full auth proof
- reading org ID from request input without middleware-backed membership resolution
- writing Casbin policy outside permission/Casbin abstractions
- moving protected route to weaker group for convenience
- forgetting API-key actor path on protected endpoints

## Stop Conditions

- stop if route ownership, tenant boundary, or auth stratum is unclear
- stop if live code contradicts cache/docs in way that changes protection decision
- stop before destructive DB/schema/data operations not explicitly requested

## Completion Output

Report:

- boundary touched
- route strata and middleware implications
- files changed
- commands run and exact result
- skipped verification and blocker
- residual risk
