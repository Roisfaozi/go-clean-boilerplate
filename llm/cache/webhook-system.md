# Webhook System

## Purpose

Durable map for webhook configuration, event subscriptions, logs, tenant scoping, and async dispatch behavior.

## Runtime truth

- Module root: `internal/modules/webhook/`
- Wiring entry: `internal/modules/webhook/module.go`
- Routes: `internal/modules/webhook/delivery/http/webhook_routes.go`
- Usecase: `internal/modules/webhook/usecase/webhook_usecase.go`
- Async task path: `internal/worker/tasks/webhook.go`
- Worker handlers and processor registration live under `internal/worker/`

## Route ownership

- Webhook routes are registered under `tenantAuthorized` routing through `webhookHttp.RegisterWebhookRoutes(tenantAuthorized, ...)`.
- This means webhook management depends on authenticated user or API-key actor, tenant context, and Casbin-protected routing.

## Behavior surfaces

- webhook CRUD
- event subscription configuration
- webhook trigger dispatch
- webhook delivery logs query
- async delivery and retry path

## Tenant and async semantics

- Webhook operations remain organization-scoped.
- Trigger dispatch is asynchronous by default and should remain so unless product intent changes.
- Any change to event names, subscriptions, or dispatch timing can affect both producers and workers.

## Coupling to other systems

- request-owned events can enqueue webhook work
- worker lifecycle and retry behavior matter
- audit or other mutation flows may rely on webhook dispatch as side effect

## Hard rules

- Do not convert async trigger behavior to sync unless explicitly intended.
- Do not weaken organization scoping on CRUD, logs, or trigger paths.
- Do not sever worker side effects without tracing downstream expectations.

## Verification and evidence paths

- `internal/modules/webhook/test/*`
- `internal/modules/webhook/usecase/webhook_usecase.go`
- `internal/modules/webhook/delivery/http/webhook_routes.go`
- `internal/worker/tasks/webhook.go`
- `internal/router/router.go`
