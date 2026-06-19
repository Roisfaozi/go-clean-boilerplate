# Role System

## Purpose

Durable map for role CRUD, validation, model conversion, role-policy cleanup, and permission integration.

## Runtime truth

- Module root: `internal/modules/role/`
- Wiring entry: `internal/modules/role/module.go`
- Routes: `internal/modules/role/delivery/http/role_routes.go`
- Controller: `internal/modules/role/delivery/http/role_controller.go`
- Usecase: `internal/modules/role/usecase/role_usecase.go`
- Repository: `internal/modules/role/repository/role_repository.go`
- Converter and validation paths: `internal/modules/role/model/*`
- `NewRoleModule` wires DB, logger, validator, transaction manager, role repository, and permission usecase.

## Route ownership

- `internal/router/router.go` registers role routes under `authorized` group via `roleHttp.RegisterAuthorizedRoutes(...)`.
- Authorized route layering includes API-key auth, JWT/session, explicit `admin:manage` API-key scope for API-key actors, user status, optional organization, and Casbin middleware.

## Behavior surfaces

- role CRUD
- role validation and normalization
- role model conversion
- role-policy cleanup or orchestration during deletes or updates

## Coupling to other systems

- role usecase depends on permission usecase for role-policy cleanup or orchestration
- role changes can therefore affect Casbin policies and effective permission behavior
- access-right semantics can indirectly affect role meaning when permissions are expanded from registry data

## Hard rules

- Role writes that affect permissions must preserve permission-usecase integration.
- Do not bypass transaction manager when role and policy state must remain consistent.
- Preserve validation and model-conversion tests when changing role fields or normalization.
- Do not treat role admin routes as self-service routes.

## Verification and evidence paths

- `internal/modules/role/usecase/*_test.go`
- `internal/modules/role/repository/role_repository_test.go`
- `internal/modules/role/delivery/http/role_controller_test.go`
- `internal/modules/role/model/*_test.go`
- `internal/router/router.go`
