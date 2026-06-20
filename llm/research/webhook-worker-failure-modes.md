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

## Confirmed Implementation Notes

- Audit `LogActivity` writes transactional request-side audit data into `audit_outbox`; non-transactional audit logs write directly to `audit_logs` and broadcast realtime events.
- Audit outbox sync now maps `AuditOutbox.ID` to `AuditLog.ID`, giving retries a stable primary key and preventing duplicate audit rows when a moved entry is replayed.
- Webhook trigger handling remains async through Asynq and does not dedupe outbound HTTP calls; retryable delivery failures now return webhook-log persistence failure too, so Asynq can retry instead of hiding observability failure.

## Remaining Verification Gate

- Root worker Redis/miniredis tests require localhost sockets; sandbox run is blocked with `listen tcp 127.0.0.1:0: socket: operation not permitted` and must be verified from a host/socket-capable shell.
