# Worker, Audit, and Webhook System

## Purpose

Durable map for async side effects, Asynq workers, audit behavior, webhook dispatch, cleanup, and email tasks.

## Runtime truth

- `internal/worker` owns distributor, processor, scheduler, handlers, and task payloads.
- `internal/config/app.go` wires task distributor/processor/scheduler and module side effects.
- `internal/modules/audit` owns audit logs and outbox behavior.
- `internal/modules/webhook` owns webhook configs/logs/dispatch usecase and repository.

## Worker lifecycle

- Usecases enqueue work through task distributor.
- Task payloads live under `internal/worker/tasks/*`.
- Processor registration lives in `internal/worker/processor.go`.
- Side-effect handlers live under `internal/worker/handlers/*`.
- Scheduler behavior lives in `internal/worker/scheduler.go`.

## Side effect domains

- audit log processing / outbox sync
- webhook dispatch
- email delivery
- cleanup jobs
- scheduled/periodic work

## Hard rules

- Do not silently convert async behavior to sync or sync behavior to async.
- Preserve retry/idempotency assumptions before changing handler writes.
- Keep audit/webhook consistency with primary request transaction behavior.
- Use integration tests when request response semantics depend on async side effects.

## Tests and evidence paths

- `internal/worker/*_test.go`
- `internal/worker/handlers/*_test.go`
- `internal/worker/tasks/*_test.go`
- `internal/modules/audit/test/*`
- `internal/modules/webhook/test/*`
