# Role System

## Purpose

Durable map for role CRUD, role validation, role-policy cleanup, and permission integration.

## Runtime truth

- Module root: `internal/modules/role/module.go`.
- Controller/routes: `internal/modules/role/delivery/http/*`.
- Usecase: `internal/modules/role/usecase/role_usecase.go`.
- Repository: `internal/modules/role/repository/role_repository.go`.
- Converter/model validation: `internal/modules/role/model/*`.
- `NewRoleModule` wires DB, logger, validator, transaction manager, role repository, and permission usecase.

## Route ownership

- `internal/router/router.go` registers role routes under `authorized` route group via `roleHttp.RegisterAuthorizedRoutes`.
- Authorized group applies API-key auth, JWT/session, explicit `admin:manage` API-key scope for API keys, user status, optional organization, and Casbin middleware.

## Dependency boundary

Role usecase depends on permission usecase for role-policy cleanup/orchestration. Role changes can therefore affect Casbin policies and permission behavior.

## Hard rules

- Role writes that affect permissions must preserve permission usecase integration.
- Do not bypass transaction manager when role and policy state must remain consistent.
- Preserve validation/model conversion tests when changing role fields.

## Tests and evidence paths

- `internal/modules/role/usecase/*_test.go`
- `internal/modules/role/repository/role_repository_test.go`
- `internal/modules/role/delivery/http/role_controller_test.go`
- `internal/modules/role/model/*_test.go`
