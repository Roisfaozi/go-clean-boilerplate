---
name: user-domain
description: Use when changing user profile, registration, status, avatar, list or search behavior, or user-related side effects in this Casbin repo, especially when auth, storage, audit, webhook, tenant, or querybuilder boundaries may be involved.
---

# User Domain

## Overview

User domain is broad and dependency-heavy.

It spans public registration, authenticated self-service, authorized admin management, avatar/storage hooks, querybuilder-sensitive list/search, and side effects into audit, auth, and webhook flows.

## When To Use

Use this skill when:

- changing registration or profile behavior
- changing user list/search or status management
- changing avatar or storage-related user flow
- changing user side effects touching audit, auth/session, webhook, or Casbin

## Read Order

1. `AGENTS.md`
2. `llm/cache/user-system.md`
3. `llm/cache/querybuilder-security.md`
4. `llm/cache/authentication-system.md`
5. `llm/cache/tenant-organization-system.md` when org context matters
6. `internal/router/router.go`
7. `internal/modules/user/module.go`
8. `internal/modules/user/delivery/http/user_routes.go`
9. `internal/modules/user/usecase/user_usecase.go`
10. `internal/modules/user/usecase/avatar_hook.go`
11. `internal/modules/user/repository/user_repository.go`

## Runtime Truth To Preserve

- user routes exist across public, authenticated, and authorized strata
- module wiring injects transaction manager, Casbin-related dependencies, audit, auth, webhook, and storage provider
- avatar behavior is tied to storage/TUS completion flow
- list/search behavior can intersect `pkg/querybuilder` restrictions

## Workflow

### Step 1 — Classify User Surface

State whether change affects:

- public registration
- self-service authenticated profile
- authorized/admin user management
- avatar/storage hook
- list/search/querybuilder path

### Step 2 — Trace Side Effects

Check whether path touches:

- auth/session behavior
- tenant or organization membership
- audit logging
- webhook triggers
- storage provider abstraction

### Step 3 — Patch Narrowly

- keep controller bind/validate/respond only
- keep user business rules in usecase
- keep repository and querybuilder restrictions explicit
- preserve context propagation for storage/avatar flows

### Step 4 — Verify

Start with:

- `internal/modules/user/test/user_usecase_test.go`
- `internal/modules/user/test/user_controller_test.go`
- `internal/modules/user/repository/user_repository_test.go`
- `internal/modules/user/test/avatar_hook_test.go` when avatar flow changes

Add broader auth or querybuilder checks when those boundaries move.

## Common Mistakes

- weakening querybuilder restrictions for user search fields
- changing avatar flow without checking storage context propagation
- mixing public and authorized user semantics in one handler path
- forgetting webhook or audit side effects after user mutations

## Stop Conditions

- stop if route owner stratum is unclear between public, authenticated, and authorized user flows
- stop if avatar/storage hook change is requested but TUS/storage boundary is not reviewed

## Completion Output

Report:

- user surface changed
- side effects or boundary implications
- files changed
- verification run and exact result
