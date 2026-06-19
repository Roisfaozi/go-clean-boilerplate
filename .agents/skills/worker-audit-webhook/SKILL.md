---
name: worker-audit-webhook
description: Use when changing Asynq worker tasks, audit outbox/sync behavior, webhook dispatch, email jobs, cleanup jobs, scheduler behavior, or request side effects that enqueue background work.
---

# Worker Audit Webhook: Async Side Effects

**Announce at start:** "I'm using the worker-audit-webhook skill to preserve async side-effect semantics."

## Read Order

1. `llm/cache/backend-map.md`
2. `llm/cache/domain-rules.md`
3. `internal/worker/tasks`
4. `internal/worker/distributor.go`
5. `internal/worker/processor.go`
6. `internal/worker/handlers/*`
7. `internal/modules/audit`
8. `internal/modules/webhook`

## Workflow

### Phase 1 — Trace Task Lifecycle

`usecase -> distributor -> task payload -> processor registration -> handler -> side effect`.

### Phase 2 — Semantics

Decide:

- sync vs async guarantee
- retry behavior
- idempotency
- transaction coupling
- audit/webhook visibility to caller

### Phase 3 — Patch Rules

- Do not silently convert sync behavior to async or async behavior to sync.
- Keep task payload version/shape compatible with handlers.
- Keep audit/webhook consistency with primary request transaction behavior.

### Phase 4 — Verify

- unit tests for task payload/handler logic
- integration tests where request response depends on async side effects
- scheduler tests where timing/cleanup changes

## Stop Conditions

- Stop and ask before destructive DB/schema/data operations not explicitly requested.
- Stop if live code contradicts `llm/cache/*`; live code wins, then document drift in `llm/tasks/`.
- Stop if route ownership, tenant boundary, or auth stratum is unclear.

## Completion Output

Report:

- files changed
- commands run and exact result
- verification skipped and exact blocker
- risks or follow-up work
