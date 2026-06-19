---
name: audit-domain
description: Use when changing audit logs, audit outbox, audit listing routes, or request side effects that emit audit data in this Casbin repo, especially when persistence, organization visibility, or worker-sync coupling must remain correct.
---

# Audit Domain

## Overview

Audit changes affect both stored visibility and side-effect delivery.

This repo uses audit log plus outbox behavior, so changes can break operator visibility or downstream sync if treated as simple CRUD.

## When To Use

Use this skill when:

- changing audit list/query behavior
- changing audit log creation or outbox writes
- changing cleanup or sync behavior for audit data
- changing request side effects that emit audit records

Do not use this skill for:

- generic worker plumbing without audit semantics; use `worker-audit-webhook`
- generic endpoint contract changes only; use `api-endpoint`

## Read Order

1. `AGENTS.md`
2. `llm/cache/audit-system.md`
3. `llm/cache/worker-audit-webhook-system.md`
4. `internal/router/router.go`
5. `internal/modules/audit/module.go`
6. `internal/modules/audit/delivery/http/audit_routes.go`
7. `internal/modules/audit/usecase/audit_usecase.go`
8. `internal/modules/audit/repository/audit_repository.go`

## Runtime Truth To Preserve

- audit module lives under `internal/modules/audit/`
- audit routes are registered through `auditHttp.RegisterAuthorizedRoutes(authorized, ...)` in `internal/router/router.go`
- repository applies organization visibility scoping when listing audit logs
- audit repository also manages outbox persistence and status transitions

## Workflow

### Step 1 — Classify Change

State whether change affects:

- listing/filter/sort visibility
- log creation
- outbox enqueue/update/delete behavior
- prune/sync/worker coupling

### Step 2 — Trace Organization Scope

Verify whether code relies on:

- request organization context
- organization visibility scope in repository queries
- authorized/admin route boundary

Do not loosen audit visibility accidentally.

### Step 3 — Patch Persistence And Side Effects

- preserve audit log shape and timestamps unless explicitly changing contract
- preserve outbox retry/status behavior when touching async sync path
- preserve request-to-audit coupling if action side effects still need visibility

### Step 4 — Verify

Start with:

- `internal/modules/audit/test/audit_usecase_test.go`
- `internal/modules/audit/test/audit_repository_test.go`
- `internal/modules/audit/test/audit_controller_test.go`

Expand to worker verification when outbox processing semantics changed.

## Common Mistakes

- changing list query without preserving organization visibility
- treating outbox as optional side table
- moving audit side effects out of transaction-aware flow without reviewing worker coupling

## Stop Conditions

- stop if change would alter audit visibility scope and requester intent is not explicit
- stop if outbox behavior changes but worker-owned consumer path is not reviewed

## Completion Output

Report:

- audit surface changed
- visibility or outbox implications
- files changed
- verification run and blocker if any
