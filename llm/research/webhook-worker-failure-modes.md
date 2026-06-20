# Webhook Worker Failure Modes

## Scope

Phase 5 audit untuk worker, audit, webhook, dan retry/idempotency.

## Evidence Paths

- `internal/modules/audit/*`
- `internal/modules/webhook/*`
- `internal/worker/*`
- `internal/worker/handlers/*`
- `internal/worker/tasks/*`

## Failure Modes

- enqueue before DB commit can leave orphaned task state
- retry without idempotency can duplicate webhook/email/audit side effects
- delivery logs can hide partial failure if caller only sees request success
- cleanup jobs can race with new writes if key selection too broad

## What to Verify

- task enqueue occurs after successful transaction or on safe outbox path
- webhook trigger handler remains idempotent on duplicate task
- audit sync path does not double-write on worker retry
- email cleanup and webhook cleanup jobs cannot delete new data by stale selector

## Needs Confirmation

- exact outbox vs direct enqueue split in current code
- which worker handlers already dedupe by event id
