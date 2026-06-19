# Webhook System

## Purpose

Durable map for webhook configuration, event trigger, logs, and worker dispatch behavior.

## Runtime truth

- `internal/modules/webhook/module.go` wires repository, worker distributor, controller, and usecase.
- `internal/modules/webhook/delivery/http/webhook_routes.go` registers webhook routes under tenantAuthorized group.
- `internal/modules/webhook/usecase/webhook_usecase.go` owns create/update/delete/find/logs/trigger behavior.
- `internal/worker/tasks/webhook.go` and webhook handler path own async execution.

## Behavior surfaces

- webhook CRUD
- event subscription lists
- webhook trigger dispatch to worker
- webhook logs query

## Hard rules

- Trigger dispatch should stay asynchronous unless product intent changes.
- Webhook operations remain organization-scoped.
- Do not sever worker side effect behavior without updating audit/trigger expectations.

## Tests and evidence paths

- `internal/modules/webhook/test/*`
- `internal/modules/webhook/usecase/webhook_usecase.go`
- `internal/modules/webhook/delivery/http/webhook_routes.go`
