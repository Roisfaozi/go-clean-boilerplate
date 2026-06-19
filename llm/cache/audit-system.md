# Audit System

## Purpose

Durable map for audit log visibility, audit outbox behavior, authorized route ownership, and side effects that publish audit data.

## Runtime truth

- Module root: `internal/modules/audit/`
- Wiring entry: `internal/modules/audit/module.go`
- Routes: `internal/modules/audit/delivery/http/audit_routes.go`
- Controller: `internal/modules/audit/delivery/http/audit_controller.go`
- Usecase: `internal/modules/audit/usecase/audit_usecase.go`
- Repository: `internal/modules/audit/repository/audit_repository.go`
- Entities include both audit log and audit outbox records.

## Route ownership

- Audit routes are registered through `auditHttp.RegisterAuthorizedRoutes(authorized, auditModule.AuditController)` in `internal/router/router.go`.
- That means audit log access is part of admin-style authorized routing, not public or tenant-self-service routing.

## Data and visibility behavior

- Audit repository uses organization visibility scoping when listing logs.
- Audit log queries support dynamic filter and sort behavior through querybuilder.
- Audit outbox records are persisted and later processed through async or follow-up paths.

## Behavior surfaces

- audit log listing and querying
- audit log creation from request-owned side effects
- audit outbox creation, retry, status update, and deletion
- pruning old audit logs
- any sync pipeline that consumes outbox rows

## Coupling to other systems

- user, auth, permission, and other mutation flows can emit audit data
- worker/audit/webhook behavior can depend on audit side effects staying explicit
- organization visibility rules matter for operator-facing log views

## Hard rules

- Audit writes that are coupled to request or worker side effects must stay explicit.
- Do not weaken organization visibility when changing list/query logic.
- Do not treat audit outbox as optional if async sync behavior still depends on it.
- Do not silently remove audit publication from flows that currently expose it.

## Verification and evidence paths

- `internal/modules/audit/test/*`
- `internal/modules/audit/repository/audit_repository.go`
- `internal/modules/audit/usecase/audit_usecase.go`
- `internal/modules/audit/delivery/http/audit_routes.go`
- `internal/router/router.go`
