# Webhook System

## Purpose

Durable map for webhook-domain behavior:

- webhook CRUD
- event subscription semantics
- tenant scoping
- webhook log behavior
- async trigger dispatch through worker

Use before changing `internal/modules/webhook`, webhook event names, subscription filtering, delivery logs, or worker webhook tasks.

## Primary source of truth

1. `internal/modules/webhook/delivery/http/webhook_routes.go`
2. `internal/modules/webhook/delivery/http/webhook_controller.go`
3. `internal/modules/webhook/usecase/webhook_usecase.go`
4. `internal/modules/webhook/repository/webhook_repository.go`
5. `internal/worker/tasks/webhook.go`
6. `internal/worker/handlers/webhook_handler.go`
7. `internal/worker/processor.go`
8. `internal/router/router.go`

## Runtime ownership

- webhook module owns config CRUD, subscription choices, trigger request construction, and delivery log persistence.
- worker owns queued dispatch execution and retry behavior.
- router owns tenant-authorized placement and API-key scope layering.

## Route ownership

Webhook routes are registered under `tenantAuthorized`.

This means webhook management requires:

- authenticated user or API-key actor
- user status check
- organization context
- Casbin authorization
- API-key scope checks when route applies explicit scopes

Webhook routes are organization-scoped, not global admin-only defaults.

## Async dispatch model

Typical flow:

`producer usecase -> webhook usecase Trigger -> task distributor -> webhook task payload -> processor -> webhook handler -> delivery log/update`

Implications:

- request may enqueue work and return before delivery completes
- retry/idempotency behavior belongs to worker path
- delivery logs must remain consistent with handler outcomes

## Event contract

Webhook event names and payload shape are external contract.

Changing event names, filtering, payload keys, or delivery timing can break consumers even if internal tests pass.

Known producer examples include user lifecycle triggers and other mutation flows.

## Cross-system coupling

Webhook changes can affect:

- worker queue and retry behavior
- audit or mutation flows that emit events
- frontend/admin webhook management screens
- tenant organization scoping
- API-key access to webhook management endpoints

## Known sharp edges

- async dispatch failure may show up only in worker logs or webhook logs.
- organization ID must stay attached through trigger and delivery path.
- event subscription filtering can silently stop deliveries.
- synchronous rewrite can create latency or failure-coupling regression.

## Change checklist

Before editing webhook code, prove:

1. Is change CRUD, trigger, worker handler, or logs?
2. Does organization ID remain scoped throughout?
3. Does event name/payload change externally?
4. Does worker retry/idempotency still make sense?
5. Does frontend/admin UI need contract update?

## Verification paths

- `internal/modules/webhook/test/*`
- `internal/modules/webhook/usecase/webhook_usecase.go`
- `internal/modules/webhook/delivery/http/webhook_routes.go`
- `internal/worker/tasks/webhook.go`
- `internal/worker/handlers/*webhook*`
- `internal/router/router.go`

## Hard rules

- Do not weaken organization scoping.
- Do not convert async trigger to sync without explicit product decision.
- Do not rename event contracts casually.
- Do not drop delivery logging or retry semantics without replacement.
