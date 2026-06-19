# Access Right System

## Purpose

Durable map for access-right registry behavior, route ownership, and the resource-action contract that feeds permission and Casbin flows.

## Runtime truth

- Module root: `internal/modules/access/`
- Wiring entry: `internal/modules/access/module.go`
- Controller and routes: `internal/modules/access/delivery/http/*`
- Usecase: `internal/modules/access/usecase/access_usecase.go`
- Repository: `internal/modules/access/repository/access_repository.go`
- Model and entity: `internal/modules/access/model/*`, `internal/modules/access/entity/*`

## Route ownership

- `internal/router/router.go` registers access routes under `authorized` group via `accessHttp.RegisterAccessRoutes(authorized.Group("", tenantMiddleware.OptionalOrganization()), accessModule.AccessController)`.
- This means access routes inherit admin-style authorized middleware and also use optional organization middleware at subgroup registration.

## Behavior surfaces

- access-right CRUD or registry updates
- endpoint or resource-action listing used by admin surfaces
- resource-action data consumed by permission assignment and effective permission expansion

## Coupling to other systems

- permission module depends on `accessModule.AccessRepo` for access-right expansion behavior
- role and user permission assignment can change meaning when access-right names or actions change
- Casbin-effective behavior can drift if access-right registry and permission mapping get out of sync

## Contract semantics

- access-right names, resources, and actions are authorization contract data, not display-only labels
- route ownership is admin-style and security-sensitive
- optional organization context on subgroup registration is part of actual runtime behavior and should not be ignored when changing access flows

## Hard rules

- Treat access-right names, resources, and actions as security contract, not UI metadata.
- Coordinate access-right changes with permission and Casbin behavior.
- Do not bypass authorized route semantics or optional organization subgroup behavior.
- Do not rename resource-action semantics casually without tracing downstream permission expansion.

## Verification and evidence paths

- `internal/modules/access/test/*`
- `internal/modules/access/repository/access_repository_test.go`
- `internal/modules/access/usecase/access_usecase.go`
- `internal/modules/permission/usecase/access_right_assignment.go`
- `internal/router/router.go`
