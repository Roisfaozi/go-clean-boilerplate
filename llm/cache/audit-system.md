# Audit System

## Purpose

Durable map for audit-domain behavior:

- audit log creation
- organization visibility scoping
- dynamic audit search
- audit outbox
- export flow
- worker/async coupling

Use before changing `internal/modules/audit`, audit side effects in other modules, audit log search/export, or worker audit processing.

## Primary source of truth

1. `internal/modules/audit/delivery/http/audit_routes.go`
2. `internal/modules/audit/delivery/http/audit_controller.go`
3. `internal/modules/audit/usecase/audit_usecase.go`
4. `internal/modules/audit/repository/audit_repository.go`
5. `internal/modules/audit/entity/audit_log.go`
6. `internal/modules/audit/entity/audit_outbox.go`
7. `internal/worker/*`
8. `internal/router/router.go`

## Runtime ownership

- audit module owns audit log persistence, outbox persistence, listing, export, and websocket/event publication behavior.
- repository owns organization visibility scopes and dynamic querybuilder use.
- worker owns async export/outbox processing where configured.

## Route ownership

Audit routes are registered under `authorized` group.

Audit access therefore inherits:

- API-key auth
- JWT/session validation
- explicit `admin:manage` for API-key actors
- user status
- optional organization context
- Casbin middleware

Audit logs are not public or tenant-self-service by default.

## Log creation semantics

`LogActivity` can behave differently by context:

- transactional path can write audit outbox
- non-transactional path can create log and distribute/notify directly
- organization context can be captured from request context

This means audit behavior is part of request transaction semantics.

## Listing and visibility

Repository dynamic list/search uses:

- `database.OrganizationScope(ctx)`
- `database.OrganizationVisibilityScope(ctx, "audit_logs.organization_id")`
- `pkg/querybuilder` for dynamic filters/sorts

Do not weaken visibility scope while changing search behavior.

## Outbox behavior

Audit outbox owns pending/processing/failed/completed states.

Repository supports:

- create outbox
- find pending outbox with retry count under limit
- update status and retry count
- delete outbox

Outbox changes must preserve retry and failure semantics.

## Cross-system coupling

Audit side effects appear in:

- user registration/profile/avatar/status/delete flows
- auth/session flows
- permission and role changes
- worker cleanup/export behavior
- websocket/realtime notification path

## Known sharp edges

- audit search can leak cross-org data if visibility scopes are weakened.
- transactional outbox behavior differs from non-transactional direct behavior.
- CSV/export paths must guard formula injection and streaming limits.
- tests may need SQL expectation updates when organization visibility scopes change.

## Change checklist

Before editing audit code, prove:

1. Is change on log create, list/search, export, or outbox?
2. Does organization visibility scope still apply?
3. Does querybuilder expose new sensitive fields?
4. Does transaction context require outbox path?
5. Does worker/retry behavior still match entity states?

## Verification paths

- `internal/modules/audit/test/*`
- `internal/modules/audit/repository/audit_repository.go`
- `internal/modules/audit/usecase/audit_usecase.go`
- `internal/modules/audit/delivery/http/audit_controller.go`
- `internal/worker/*audit*`
- `internal/router/router.go`

## Hard rules

- Do not weaken organization visibility scoping.
- Do not silently remove audit side effects from mutating flows.
- Preserve outbox retry/status semantics.
- Treat dynamic audit filters as querybuilder security surface.
