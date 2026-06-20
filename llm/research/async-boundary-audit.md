# Async Boundary Audit

## Scope

Audit ini memetakan boundary worker, audit, webhook, upload completion, SSE/WS presence, dan retry/idempotency risk.

## Evidence Paths

- `internal/modules/audit/*`
- `internal/modules/webhook/*`
- `internal/worker/*`
- `internal/worker/handlers/*`
- `internal/worker/tasks/*`
- `pkg/tus/*`
- `pkg/ws/*`
- `pkg/sse/*`

## Verified Facts

- Worker handles audit, webhook, cleanup, and email-related tasks.
- Webhook dispatch is async.
- SSE and WS are stateful, Redis-backed in parts, and depend on context/ticket validation.
- TUS completion can mutate domain state such as user avatar.

## Weaknesses

- Request success can precede side-effect completion.
- Retry semantics can duplicate work if handlers are not idempotent.
- Presence/broadcast cleanup can drift in distributed realtime.
- Upload completion can leave orphaned storage if domain mutation fails.

## Recommendations

- document enqueue timing relative to DB commit
- add idempotency tests for worker handlers
- add cleanup tests for upload completion failures
- add race-focused tests for WS/presence paths

## Needs Confirmation

- which worker tasks are idempotent today
- which upload hooks mutate domain state beyond avatar
- whether realtime presence cleanup is fully covered by E2E
