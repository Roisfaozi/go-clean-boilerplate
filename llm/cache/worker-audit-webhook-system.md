# Worker, Audit, and Webhook System

## Purpose

Durable map for async side-effect behavior:

- Asynq distributor and processor lifecycle
- scheduled cleanup jobs
- audit export/outbox processing
- webhook dispatch
- email delivery
- request paths that enqueue background tasks

Use before changing `internal/worker`, audit/webhook side effects, cleanup jobs, task payloads, or task retry behavior.

## Primary source of truth

1. `internal/config/app.go`
2. `internal/worker/distributor.go`
3. `internal/worker/processor.go`
4. `internal/worker/scheduler.go`
5. `internal/worker/tasks/*`
6. `internal/worker/handlers/*`
7. `internal/modules/audit/*`
8. `internal/modules/webhook/*`

## Runtime ownership

`internal/config/app.go` wires:

- Redis Asynq client options
- task distributor
- cleanup handler
- webhook handler
- audit usecase/repository into task processor
- SMTP config into worker config
- scheduler and scheduled tasks
- processor start/shutdown lifecycle

This file is the truth for which worker dependencies exist at runtime.

## Async path model

Common shape:

`usecase -> task distributor -> serialized payload -> processor registration -> handler -> external/database side effect -> retry/log outcome`

Any change to payload schema must update distributor, task definition, processor, handler, and tests together.

## Side-effect domains

Current worker-relevant domains:

- webhook trigger dispatch
- email sending
- audit log export/outbox sync
- cleanup jobs for stale or expired data
- scheduled periodic tasks

## Transaction and timing semantics

- enqueuing inside request flow does not mean side effect completed before response.
- task enqueue before DB commit can create race if handler reads uncommitted/missing state.
- converting async to sync changes latency and failure coupling.
- converting sync to async changes consistency guarantees and user-visible timing.

## Retry and idempotency

Worker changes must preserve:

- retry count behavior
- safe repeated delivery where possible
- durable status/log updates for audit/webhook work
- clear failure logging

If handler can run twice, write path must tolerate duplicate effects or guard them.

## Cross-system coupling

Worker changes can break:

- webhook module delivery guarantees
- audit outbox/export behavior
- auth/user email flows
- cleanup retention assumptions
- app shutdown behavior

Read with:

- `llm/cache/audit-system.md`
- `llm/cache/webhook-system.md`
- `llm/cache/authentication-system.md`

## Known sharp edges

- worker starts only when app startup reaches processor start path.
- scheduler shutdown must pair with processor shutdown.
- SMTP config is injected through app wiring, not handler globals.
- tests may pass at handler level but fail integration due Redis/Asynq config.

## Change checklist

Before editing worker code, prove:

1. Which producer enqueues task?
2. Which payload struct is serialized?
3. Where is task registered in processor?
4. Which handler performs side effect?
5. What retry/idempotency behavior is expected?
6. Does request assume side effect completion or eventual completion?

## Verification paths

- `internal/worker/*_test.go`
- `internal/worker/handlers/*_test.go`
- `internal/worker/tasks/*_test.go`
- `internal/modules/audit/test/*`
- `internal/modules/webhook/test/*`
- `internal/config/app.go`

## Hard rules

- Do not silently change async/sync behavior.
- Do not change task payload without updating producer, processor, handler, and tests.
- Preserve retry/idempotency assumptions.
- Use integration verification when request semantics depend on worker execution.
