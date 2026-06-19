---
name: execute
description: Execute an approved Casbin implementation plan task-by-task with progress tracking and verification. Use after a plan exists.
---

# Execute: Plan Implementation

**Announce at start:** "I'm using the execute skill to implement the approved plan task by task."

## Prerequisites

- Implementation plan exists in `llm/tasks/todo.md` or `llm/plans/*`.
- Route/module ownership is clear.
- High-risk skill selected if needed.

## Workflow

### Step 1 — Load Plan

- Read active plan.
- Check dependencies and blockers.
- Confirm changed layers.

### Step 2 — Select Skills

Use exactly relevant skills:

- backend logic: `go-service`
- route/API: `api-endpoint`
- auth/tenant/Casbin: `auth-tenant-casbin`
- API key: `api-key-scope`
- DB transaction/migration: `database-transactions`
- upload: `tus-upload-storage`
- worker/audit/webhook: `worker-audit-webhook`
- querybuilder: `query-builder-security`
- realtime: `realtime-sse-websocket`
- frontend ownership: `frontend-surface`

### Step 3 — Implement Task

- Do one coherent slice at a time.
- Do not mix unrelated cleanup.
- Update `llm/tasks/todo.md` for multi-step progress.

### Step 4 — Verify Task

- Run narrow verification for changed layer.
- Record exact failures/blockers.
- Stop and re-plan if root assumption is wrong.

### Step 5 — Complete

- Run `self-review`.
- Run `verification-before-completion`.
- Report final file list and verification.
