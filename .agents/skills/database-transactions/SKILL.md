---
name: database-transactions
description: Use when changing GORM transactions, all-or-nothing writes, schema-sensitive persistence flows, seed-like data coupling, or Casbin policy writes tied to database state in this Casbin repo.
---

# Database Transactions

## Overview

Transaction work in this repo is boundary work.

It often intersects tenant scoping, audit or webhook side effects, and permission or Casbin writes that must stay atomic with DB state.

## When To Use

Use this skill when:

- multiple writes must succeed or fail together
- usecase now spans repository plus side-effect or policy write
- GORM transaction manager or context-carried DB is involved
- schema-sensitive persistence logic changes behavior

## Read Order

1. `AGENTS.md`
2. `llm/workflows/database-migration.md` if schema shape or migration is involved
3. `llm/cache/casbin-permission-system.md` when policy writes are coupled
4. `llm/cache/worker-audit-webhook-system.md` when side effects depend on request transaction
5. target module usecase and repository
6. `pkg/tx/*` and context DB helpers when transaction manager is involved

## Runtime Truth To Preserve

- repositories may read transaction-scoped DB from context via `pkg/tx`
- policy writes tied to DB state should use transactional enforcer patterns
- async side effects must stay consistent with primary request transaction behavior

## Workflow

### Step 1 — Identify Atomic Boundary

State exactly what must commit together:

- DB rows only
- DB rows plus Casbin policy
- DB rows plus audit or webhook enqueue behavior

### Step 2 — Trace Current Transaction Owner

Find whether transaction starts in:

- usecase
- helper wrapper
- repository context path

Do not create nested or duplicate transaction ownership casually.

### Step 3 — Patch With Context Integrity

- preserve context propagation into repository methods
- keep side effects ordered relative to commit semantics
- avoid doing external irreversible work before atomic DB state is safe

### Step 4 — Verify Failure Semantics

Test not only success path, but also rollback or partial-failure path when practical.

## Common Mistakes

- writing DB and Casbin policy outside shared transaction boundary
- enqueueing async side effect before DB state is durable
- bypassing transaction-aware DB retrieval from context

## Stop Conditions

- stop if atomic ownership is unclear
- stop if change could leave DB state and policy or side effects inconsistent
- stop before destructive schema or data operations not explicitly requested

## Completion Output

Report:

- atomic boundary changed
- transaction owner path
- files changed
- verification run and exact result
- rollback-risk notes
