# Access Right System

## Purpose

Durable map for access-right and endpoint registry behavior that feeds permission/Casbin flows.

## Runtime truth

- Module root: `internal/modules/access/module.go`.
- Controller/routes: `internal/modules/access/delivery/http/*`.
- Usecase: `internal/modules/access/usecase/access_usecase.go`.
- Repository: `internal/modules/access/repository/access_repository.go`.
- Model/entity: `internal/modules/access/model` and `internal/modules/access/entity`.

## Route ownership

- `internal/router/router.go` registers access routes under `authorized` group via `accessHttp.RegisterAccessRoutes(authorized.Group("", tenantMiddleware.OptionalOrganization()), accessModule.AccessController)`.
- This means access routes inherit authorized group middleware and additionally use optional organization middleware at subgroup registration.

## Relationship to permission system

- Permission module depends on `accessModule.AccessRepo` for access-right expansion and permission behavior.
- Access-right changes can affect role/user permission assignment, batch checks, and Casbin policy materialization.

## Hard rules

- Treat access-right names/resources/actions as authorization contract, not display-only data.
- Coordinate access-right changes with `llm/cache/casbin-permission-system.md`.
- Do not bypass authorized group or optional organization semantics.

## Tests and evidence paths

- `internal/modules/access/test/*`
- `internal/modules/access/repository/access_repository_test.go`
- `internal/modules/permission/usecase/access_right_assignment.go`
- `internal/router/router.go`
