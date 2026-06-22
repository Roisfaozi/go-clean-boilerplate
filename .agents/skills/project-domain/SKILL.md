---
name: project-domain
description: Use when changing tenant-scoped project CRUD, project route protection, or project-specific API-key scope behavior in this Casbin repo.
---

# Project Domain

## Overview

Project endpoints are tenant-scoped and explicitly API-key scoped.

Project changes must preserve organization context and keep route-level protection aligned with product intent.

## When To Use

Use this skill when:

- changing project CRUD logic or payloads
- changing project route protection
- changing project-specific API-key scopes
- changing tenant-specific project visibility

## Read Order

1. `AGENTS.md`
2. `llm/cache/project-system.md`
3. `llm/cache/api-key-system.md`
4. `llm/cache/tenant-organization-system.md`
5. `internal/router/router.go`
6. `internal/modules/project/module.go`
7. `internal/modules/project/usecase/project_usecase.go`
8. `internal/modules/project/repository/project_repository.go`
9. `internal/modules/project/delivery/http/project_controller.go`

## Runtime Truth To Preserve

- project routes are defined directly in `internal/router/router.go`
- they live under `tenantAuthorized` group
- route scopes are explicit:
  - `project:manage` for create, update, delete
  - `project:view` or `project:manage` for list and detail
- tenant organization context is required before project handlers run

## Workflow

### Step 1 — Confirm Tenant Boundary

Verify project flow still depends on:

- authenticated user or API-key actor
- tenant middleware organization context
- Casbin-protected tenant route boundary

### Step 2 — Confirm Scope Matrix

Check whether change affects any of:

- `project:view`
- `project:manage`
- list vs detail vs mutation differences

### Step 3 — Patch Usecase And Repository Narrowly

- keep controller bind/validate/respond only
- keep tenant-specific rules in usecase/repository flow
- keep route-level scope checks explicit in router if route set changes

### Step 4 — Verify

Start with:

- `internal/modules/project/test/project_usecase_test.go`

Add route or integration verification when scope or tenant path changes.

## Common Mistakes

- moving project security decisions into handler body
- forgetting explicit scope strings when adding new project endpoint
- weakening tenant requirement for list or detail path

## Stop Conditions

- stop if route should move out of `tenantAuthorized` but product reason is not explicit
- stop if project visibility change depends on org semantics not yet traced in usecase/repository

## Completion Output

Report:

- project route or logic changed
- tenant and scope implications
- files changed
- verification run and exact result
