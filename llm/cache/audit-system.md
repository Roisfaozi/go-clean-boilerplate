# Audit System

## Purpose

Durable map for audit logs, audit outbox behavior, and controller/usecase/repository flow.

## Runtime truth

- `internal/modules/audit/module.go` wires repository/usecase/controller.
- `internal/modules/audit/delivery/http/audit_routes.go` registers authorized audit routes.
- `internal/modules/audit/usecase/audit_usecase.go` owns audit behavior and any async/sync flow decisions.
- `internal/modules/audit/repository/audit_repository.go` owns persistence.

## Behavior surfaces

- audit log listing/querying
- audit outbox sync/processing
- write-following side effects from user/auth/permission flows

## Hard rules

- Audit writes can be coupled to worker or request side effects; keep that explicit.
- Do not remove audit visibility from flows that currently publish it.

## Tests and evidence paths

- `internal/modules/audit/test/*`
- `internal/modules/audit/usecase/audit_usecase.go`
- `internal/modules/audit/delivery/http/audit_routes.go`
