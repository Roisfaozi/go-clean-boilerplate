---
name: access-domain
description: Use when changing access-right registry, access endpoint definitions, or permission-expansion inputs in this Casbin repo, especially when resource-action semantics must stay aligned with permission, role, and Casbin behavior.
---

# Access Domain

## Overview

This skill protects access-right registry semantics.

Access changes are rarely isolated. They feed permission expansion, role orchestration, and authorized-route access management.

## When To Use

Use this skill when:

- adding or changing access resources or actions
- changing access CRUD routes or payloads
- changing access data consumed by permission assignment logic

Do not use this skill for:

- pure route protection changes with no registry effect; use `api-endpoint`
- pure permission assignment or Casbin write logic; use `permission-domain`

## Read Order

1. `AGENTS.md`
2. `llm/cache/access-right-system.md`
3. `llm/cache/casbin-permission-system.md`
4. `llm/cache/permission-system.md`
5. `internal/router/router.go`
6. `internal/modules/access/module.go`
7. `internal/modules/access/delivery/http/access_routes.go`
8. `internal/modules/access/usecase/access_usecase.go`
9. `internal/modules/access/repository/access_repository.go`
10. `internal/modules/permission/usecase/access_right_assignment.go`

## Runtime Truth To Preserve

- access module lives under `internal/modules/access/`
- access routes are registered through `accessHttp.RegisterAccessRoutes(...)` inside authorized routing in `internal/router/router.go`
- router passes `authorized.Group("", tenantMiddleware.OptionalOrganization())`, so access endpoints sit in admin-style authorized flow with optional organization context
- access registry changes can alter permission expansion inputs used by permission usecase

## Workflow

### Step 1 — Classify Registry Change

State whether change affects:

- resource names
- action names
- access CRUD contract
- expansion inputs consumed by permission assignment

### Step 2 — Trace Consumers

Check who consumes access-right data:

- access controller and usecase
- permission assignment logic in `internal/modules/permission/usecase/access_right_assignment.go`
- role and permission admin flows if resource-action matrix is exposed there

### Step 3 — Patch Narrowly

- keep controller limited to bind/validate/respond
- keep registry rules in usecase or repository layer where they already live
- preserve naming semantics expected by permission and Casbin flows

### Step 4 — Reconfirm Route Boundary

If route shape changes, verify:

- route remains in authorized/admin-style area unless product intent explicitly changes
- optional organization context is still intentional
- API behavior still matches admin ownership expectations

### Step 5 — Verify

Start with:

- `internal/modules/access/repository/access_repository_test.go`
- `internal/modules/access/test/access_usecase_test.go`
- `internal/modules/access/test/access_controller_test.go`

Add permission-side verification when expansion behavior changes:

- `internal/modules/permission/test/access_right_assignment_test.go`
- `internal/modules/permission/test/permission_usecase_test.go`

## Common Mistakes

- renaming resource/action values without checking permission expansion consumers
- treating access registry as standalone CRUD with no policy impact
- moving registry semantics into handler/controller

## Stop Conditions

- stop if resource-action naming change would ripple into permission or role behavior not yet traced
- stop if route protection change is requested but owner boundary is unclear between access and permission domains

## Completion Output

Report:

- registry semantics changed
- consumer paths checked
- files changed
- verification run and exact result
