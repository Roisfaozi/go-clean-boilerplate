# Tenant Organization System

## Purpose

Durable map for organization tenant boundary, membership, invitations, cached reader behavior, and organization context layering on protected routes.

## Runtime truth

- `internal/middleware/tenant_middleware.go` owns required and optional organization context at HTTP middleware boundary.
- `internal/modules/organization` owns organization lifecycle, membership, invitations, cached reader behavior, restore or hard-delete style flows, and tenant-sensitive usecases.
- `internal/router/router.go` applies tenant middleware before Casbin on tenant-authorized routes.
- `internal/config/app.go` wires organization module with DB, Redis, task distributor, user repo, transaction manager, enforcer, presence manager, and frontend base URL.

## Route layering

- `tenantAuthorized`
  - API-key auth
  - JWT or session validation
  - API-key auto scope
  - user status
  - required organization
  - Casbin
- `authorized`
  - API-key auth
  - JWT or session validation
  - explicit admin scope
  - user status
  - optional organization
  - Casbin

Organization routes also exist across public, authenticated, tenant, and admin strata.

## Organization context behavior

- Required organization context should be established before tenant-sensitive controller logic runs.
- Optional organization context on admin-style routes can still matter for domain scoping, but route does not require tenant membership in same way as `tenantAuthorized`.
- Controllers should not replace tenant resolution with ad hoc query or body fields.

## Behavior surfaces

- organization CRUD and lifecycle
- membership changes
- invitations and acceptance
- cached organization reader behavior
- organization restore, delete, and visibility flows

## Coupling to other systems

- auth establishes user identity before tenant resolution.
- Casbin enforcement depends on correct organization context ordering.
- membership changes may require cache invalidation.
- user, project, API-key, and webhook flows can all depend on organization context or membership rules.

## Hard rules

- Tenant-protected routes require organization context before Casbin authorization.
- Preserve cache invalidation or refresh behavior when membership or organization state changes.
- Invitation acceptance and member changes must respect owner, admin, and member constraints in usecases.
- Do not replace tenant checks with ad hoc request parameters in handlers.

## Verification and evidence paths

- `internal/middleware/tenant_middleware.go`
- `internal/modules/organization/usecase/*`
- `internal/modules/organization/repository/*`
- `internal/modules/organization/test/*`
- `internal/router/router.go`
- `internal/config/app.go`
