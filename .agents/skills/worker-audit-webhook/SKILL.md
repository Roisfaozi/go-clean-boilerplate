---
name: worker-audit-webhook
description: Use when changing Asynq worker tasks, audit outbox or sync behavior, webhook dispatch, email jobs, cleanup jobs, scheduler behavior, or request side effects that enqueue background work in this Casbin repo.
---

# Worker Audit Webhook

## Overview

Async side effects in this repo span request path, task payload, processor registration, and handler semantics.

These changes can silently alter durability, retry, idempotency, or user-visible behavior.

## Read Order

1. `AGENTS.md`
2. `llm/cache/worker-audit-webhook-system.md`
3. `llm/cache/audit-system.md` or `llm/cache/webhook-system.md` when relevant
4. `internal/config/app.go` if wiring or scheduler matters
5. `internal/worker/tasks/*`
6. `internal/worker/distributor.go`
7. `internal/worker/processor.go`
8. `internal/worker/handlers/*`
9. affected usecase producing task

## Workflow

### Step 1 — Trace Task Lifecycle

Trace exact path:

`usecase -> distributor -> task payload -> processor registration -> handler -> side effect`

### Step 2 — Clarify Semantics

Decide and preserve:

- sync vs async behavior
- retry expectations
- idempotency assumptions
- transaction coupling
- what caller can observe immediately vs later

### Step 3 — Patch Carefully

- do not silently convert sync to async or async to sync
- keep payload and handler shape compatible
- keep audit/webhook consistency with primary request transaction behavior

### Step 4 — Verify

- task payload or handler unit tests
- scheduler tests when timing changed
- broader integration when request semantics depend on async side effects

## Common Mistakes

- enqueueing task before DB state is durable when semantics require atomic behavior
- changing payload shape without checking handler registration and tests
- assuming async means user-visible semantics do not matter

## Stop Conditions

- stop if task lifecycle owner is unclear
- stop if transaction boundary and async enqueue ordering are not traced

## Completion Output

Report:

- task lifecycle changed
- async semantics preserved or intentionally changed
- files changed
- verification run and exact result
