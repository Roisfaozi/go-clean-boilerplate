# Tenant Organization System

## Purpose

Durable map for organization tenant boundary, membership, invitations, and tenant context.

## Runtime truth

- `internal/middleware/tenant_middleware.go` owns required/optional organization context at HTTP middleware boundary.
- `internal/modules/organization` owns organization lifecycle, members, invitations, cached reader behavior, restore/hard-delete style flows, and tenant-owned usecases.
- `internal/router/router.go` applies tenant middleware before Casbin on tenant-authorized routes.
- `internal/config/app.go` wires organization repository/usecase with DB, Redis, task distributor, user repo, transaction manager, enforcer, presence manager, and frontend base URL.

## Route layering

- `tenantAuthorized` group: API key auth -> JWT/session -> API-key auto scope -> user status -> required org -> Casbin.
- `authorized` group: API key auth -> JWT/session -> explicit admin scope -> user status -> optional org -> Casbin.
- Organization module also has public/authenticated/admin/tenant route registrations.

## Tenant behavior rules

- Tenant-protected routes require organization context before Casbin authorization.
- Organization and membership changes must preserve cache invalidation behavior.
- Invitation acceptance and member changes must respect owner/admin/member rules in usecases.
- Tenant checks should not be replaced with ad hoc request/query parameter checks in controllers.

## Tests and evidence paths

- `internal/middleware/tenant_middleware.go`
- `internal/modules/organization/usecase/*`
- `internal/modules/organization/repository/*`
- `internal/modules/organization/test/*`
- `internal/router/router.go`
