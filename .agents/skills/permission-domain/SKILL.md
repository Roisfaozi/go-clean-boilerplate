---
name: permission-domain
description: Use when changing permission policy CRUD, role or user assignment, batch permission checks, access-right expansion, or transactional Casbin behavior in this Casbin repo.
---

# Permission Domain

## Overview

Permission domain is central policy orchestration.

It bridges access-right registry, role/user assignments, batch checks, and transactional enforcer behavior. Wrong patch here can silently widen or break authorization.

## When To Use

Use this skill when:

- changing permission CRUD or assignment endpoints
- changing batch permission check behavior
- changing access-right expansion or inheritance tree logic
- changing transactional Casbin writes or rollback behavior

Do not use this skill for:

- access registry semantics only; use `access-domain`
- role model conversion only; use `role-domain`

## Read Order

1. `AGENTS.md`
2. `llm/cache/permission-system.md`
3. `llm/cache/access-right-system.md`
4. `llm/cache/casbin-permission-system.md`
5. `internal/router/router.go`
6. `internal/modules/permission/module.go`
7. `internal/modules/permission/delivery/http/permission_routes.go`
8. `internal/modules/permission/usecase/permission_usecase.go`
9. `internal/modules/permission/usecase/access_right_assignment.go`
10. `internal/modules/permission/usecase/inheritance_tree.go`
11. `internal/modules/permission/usecase/transactional_enforcer.go`

## Runtime Truth To Preserve

- batch permission check route is registered under `authenticated` group via `permissionHttp.RegisterBatchCheckRoute(...)`
- full permission admin routes are registered under `authorized` group via `permissionHttp.RegisterPermissionRoutes(...)`
- permission logic depends on access-right expansion and transactional enforcer behavior

## Workflow

### Step 1 — Classify Change

State exact concern:

- policy CRUD
- role assignment
- user assignment
- batch check
- expansion/inheritance
- transactional enforcer

### Step 2 — Trace Security Boundary

Confirm:

- whether route belongs to authenticated self-service batch check or admin-style authorized policy management
- whether organization context matters in evaluation or write path
- whether API-key scope and route middleware still match intent

### Step 3 — Patch Transactionally

- preserve controller thinness
- keep security decisions in usecase
- when write path affects Casbin and DB state together, preserve transactional behavior
- confirm access-right aggregation and inheritance logic still produce expected effective permissions

### Step 4 — Verify

Start with:

- `internal/modules/permission/test/permission_usecase_test.go`
- `internal/modules/permission/test/permission_security_test.go`
- `internal/modules/permission/test/permission_controller_test.go`
- `internal/modules/permission/test/access_right_assignment_test.go`

Run broader checks if transactional enforcer behavior changed.

## Common Mistakes

- mixing authenticated batch-check semantics with admin policy management semantics
- changing access-right expansion without rechecking inheritance behavior
- breaking DB-plus-Casbin atomicity

## Stop Conditions

- stop if write path touches both policy state and DB state but transaction boundary is unclear
- stop if batch-check path and admin path start to diverge without explicit product reason

## Completion Output

Report:

- permission surface changed
- route group and transaction implications
- files changed
- verification run and exact result
