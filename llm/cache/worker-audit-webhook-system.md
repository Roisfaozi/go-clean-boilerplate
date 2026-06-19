# Worker, Audit, and Webhook System

## Purpose

Durable map for async side effects, Asynq worker lifecycle, audit behavior, webhook dispatch, cleanup jobs, and any request path that enqueues background work.

## Runtime truth

- `internal/worker` owns distributor, processor, scheduler, handlers, and task payloads.
- `internal/config/app.go` wires task distributor, processor, scheduler, and side-effect dependencies.
- `internal/modules/audit` owns audit logs and audit outbox behavior.
- `internal/modules/webhook` owns webhook configs, logs, dispatch usecase, and repository.

## Worker lifecycle

Typical async path is:

`usecase -> task distributor -> task payload -> processor registration -> handler -> side effect`

Key areas:

- task payloads under `internal/worker/tasks/*`
- processor registration in `internal/worker/processor.go`
- side-effect handlers under `internal/worker/handlers/*`
- scheduler behavior in `internal/worker/scheduler.go`

## Side effect domains

- audit log processing or outbox sync
- webhook dispatch
- email delivery
- cleanup jobs
- scheduled or periodic work

## Coupling semantics

- request-owned usecases can enqueue work through task distributor
- async behavior can still affect user-visible semantics if request assumes eventual side effect
- transaction ordering matters when DB state and enqueue behavior must stay consistent

## Hard rules

- Do not silently convert async behavior to sync or sync behavior to async.
- Preserve retry and idempotency assumptions before changing handler writes.
- Keep audit and webhook consistency aligned with request transaction behavior.
- Use integration verification when request semantics depend on async side effects.

## Verification and evidence paths

- `internal/worker/*_test.go`
- `internal/worker/handlers/*_test.go`
- `internal/worker/tasks/*_test.go`
- `internal/modules/audit/test/*`
- `internal/modules/webhook/test/*`
- `internal/config/app.go`
