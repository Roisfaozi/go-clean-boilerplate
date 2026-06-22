---
name: api-key-scope
description: Use when changing protected endpoints, API-key authentication, scope requirements, organization-scoped API-key behavior, or route access decisions involving API keys in this Casbin repo.
---

# API Key Scope

## Overview

API-key behavior in this repo is layered into router and middleware, not controller hacks.

Protected route changes must make an explicit decision about API-key participation, auto-scope, explicit scope, and organization scope.

## When To Use

Use this skill when:

- adding or changing protected endpoints
- changing API-key auth flow or middleware behavior
- changing explicit scope strings or auto-scope behavior
- changing organization-scoped API-key access

## Read Order

1. `AGENTS.md`
2. `llm/cache/api-key-system.md`
3. `llm/cache/domain-rules.md`
4. `llm/workflows/api-endpoint.md`
5. `internal/router/router.go`
6. `internal/middleware/api_key_middleware.go`
7. `internal/middleware/auth_middleware.go`
8. `internal/middleware/tenant_middleware.go` when org-scoped behavior matters
9. `internal/modules/api_key/`
10. affected controller or route tests

## Runtime Truth To Preserve

- `authenticated` and `tenantAuthorized` groups use `Authenticate()` then `RequireScopeAuto()`
- `authenticated` also uses `RequireUserSession()`
- `authorized` uses explicit `RequireScopes("admin:manage")`
- project routes under `tenantAuthorized` add explicit `project:view` and `project:manage` scopes on top of base middleware

## Scope Decision Matrix

For every changed endpoint decide one of:

- no API-key access allowed
- API-key access allowed with auto scope only
- API-key access allowed with explicit scope requirement
- API-key access allowed with organization-sensitive tenant behavior

Document that decision before patching.

## Workflow

### Step 1 — Identify Route Group

Classify route as:

- `authenticated`
- `tenantAuthorized`
- `authorized`
- special route outside those groups

### Step 2 — Decide Scope Semantics

Confirm:

- whether auto scope is enough
- whether explicit `RequireScopes(...)` is needed
- whether user-session requirement should still apply
- whether organization context should constrain key usage

### Step 3 — Patch At Router Or Middleware Boundary

- keep API-key checks in router/middleware
- preserve identity vs scope separation
- preserve org-scoped behavior on tenant routes
- avoid ad hoc scope checks in controller unless middleware truly cannot express rule

### Step 4 — Verify Negative And Positive Paths

Check at least:

- allowed key
- missing key or token
- wrong scope
- wrong organization when route is tenant-scoped
- user-session mismatch when route still depends on user session

## Common Mistakes

- accepting API-key identity without scope validation
- adding protected route without making explicit scope decision
- checking scope ad hoc inside handler while middleware already owns boundary
- forgetting that `authenticated` and `tenantAuthorized` have different session and org semantics

## Stop Conditions

- stop if route should accept API-key actor but user-session requirement is unclear
- stop if org-scoped API-key behavior is required but tenant boundary path is not traced

## Completion Output

Report:

- route group and scope decision
- files changed
- verification matrix run
- skipped checks and blocker
