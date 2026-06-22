---
name: role-domain
description: Use when changing role CRUD, role validation, model conversion, role-policy cleanup, or permission orchestration tied to roles in this Casbin repo.
---

# Role Domain

## Overview

Role changes are not only CRUD.

They can affect validation, model conversion, permission cleanup, and downstream Casbin or assignment semantics.

## When To Use

Use this skill when:

- changing role create/update/delete behavior
- changing role validation rules or converters
- changing role-policy cleanup behavior
- changing how role flows interact with permission orchestration

## Read Order

1. `AGENTS.md`
2. `llm/cache/role-system.md`
3. `llm/cache/casbin-permission-system.md`
4. `llm/cache/access-right-system.md`
5. `internal/router/router.go`
6. `internal/modules/role/module.go`
7. `internal/modules/role/delivery/http/role_routes.go`
8. `internal/modules/role/usecase/role_usecase.go`
9. `internal/modules/role/model/converter/converter.go`
10. `internal/modules/permission/usecase/permission_usecase.go`

## Runtime Truth To Preserve

- role routes are registered under `authorized` group via `roleHttp.RegisterAuthorizedRoutes(...)`
- role module has dedicated converter and validation tests
- role changes can require permission cleanup or orchestration review

## Workflow

### Step 1 — Classify Change

State whether change affects:

- role CRUD
- validation or normalization
- conversion between model/entity/request shapes
- cleanup of attached policies or assignments

### Step 2 — Cross-Module Review

Check whether role change affects:

- permission assignment or cleanup flows
- access-right semantics
- admin-only route expectations

### Step 3 — Patch Narrowly

- keep controller thin
- keep role invariants in usecase/model conversion layer
- preserve permission cleanup where role deletion or rename impacts policy state

### Step 4 — Verify

Start with:

- `internal/modules/role/model/converter/converter_test.go`
- `internal/modules/role/model/role_validation_test.go`
- `internal/modules/role/repository/role_repository_test.go`
- `internal/modules/role/usecase/role_usecase_test.go`
- `internal/modules/role/usecase/role_security_test.go`

## Common Mistakes

- changing converter or validation in one place only
- deleting role semantics without checking policy cleanup path
- treating role admin routes as self-service routes

## Stop Conditions

- stop if role mutation implies permission cleanup but cleanup owner path is unclear
- stop if validation change could break persisted role data shape and migration path is not explicit

## Completion Output

Report:

- role concern changed
- dependent permission implications
- files changed
- verification run and exact result
