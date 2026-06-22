---
name: execute
description: Use when an approved Casbin implementation plan already exists and must be executed slice by slice with repo-accurate skill routing, disciplined scope control, and truthful verification.
---

# Execute

## Overview

This skill executes approved work without losing architecture boundaries.

One slice at a time. No opportunistic cleanup. No claiming done without real verification.

## Prerequisites

- approved plan exists in `llm/tasks/todo.md` or `llm/plans/*`
- target module or route ownership is clear
- required domain skill can be selected from current scope

## Read Order

1. active plan
2. `AGENTS.md`
3. relevant `llm/cache/*`
4. relevant `llm/workflows/*`
5. live code files for current slice only

## Skill Routing Matrix

Use exact relevant skills during implementation:

- backend module logic: `go-service`
- route and contract: `api-endpoint`
- auth, tenant, Casbin: `auth-tenant-casbin`
- API key: `api-key-scope`
- DB transaction or migration: `database-transactions`
- upload and storage: `tus-upload-storage`
- worker, audit, webhook: `worker-audit-webhook`
- querybuilder security: `query-builder-security`
- realtime: `realtime-sse-websocket`
- frontend surface ownership: `frontend-surface`
- user, role, project, access, stats, audit, webhook domain changes: matching domain skill

## Execution Workflow

### Step 1 — Load Current Slice

Before editing, restate:

- current slice goal
- exact files expected to change
- exact risks for this slice
- exact verification target for this slice

### Step 2 — Reconfirm Live Truth

For current slice, re-read only live paths that matter:

- route layer in `internal/router/router.go`
- wiring in `internal/config/app.go`
- target module constructor and boundaries
- frontend proxies and shared types if contract is touched

### Step 3 — Patch Narrowly

- change only files needed for current slice
- keep handler, usecase, repository, middleware boundaries intact
- avoid unrelated refactors and formatting churn
- update docs or task notes only when current slice truly changes them

### Step 4 — Verify Narrowly

Run smallest real check that proves current slice:

- package-specific Go tests
- app or package typecheck
- route contract or proxy checks
- integration/E2E only if boundary changed enough to require it

If blocked, record exact blocker and stop pretending coverage exists.

### Step 5 — Advance Or Replan

Advance when slice goal and verification are both satisfied.

Re-plan if:

- live code contradicts plan assumption
- blast radius is larger than planned
- module ownership changes
- new approval gate appears

## Common Mistakes

- implementing multiple slices in one pass
- skipping domain skill for risky boundary
- claiming route behavior changed without checking router/proxy/type consumer paths
- relying on cached assumptions after code contradicts them

## Stop Conditions

- stop before destructive DB/schema/data operations not explicitly requested
- stop if auth stratum, tenant boundary, or Casbin/API-key enforcement path is unclear
- stop if plan says one owner but live code proves another

## Completion Output

Report:

- slice completed
- files changed
- commands run and exact result
- skipped verification and blocker
- next slice or re-plan need
