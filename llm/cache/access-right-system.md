# Access Right System

## Purpose

Durable map for access-right registry behavior:

- endpoint/resource-action catalog
- admin access-route ownership
- permission expansion input data
- resource/action semantics that feed effective Casbin behavior

Use before changing `internal/modules/access`, access-right seed/registry data, permission assignment inputs, or API contracts that expose available access rights.

## Primary source of truth

1. `internal/modules/access/delivery/http/*`
2. `internal/modules/access/usecase/access_usecase.go`
3. `internal/modules/access/repository/access_repository.go`
4. `internal/modules/access/entity/*`
5. `internal/modules/access/model/*`
6. `internal/modules/permission/usecase/access_right_assignment.go`
7. `internal/router/router.go`

## Runtime ownership

- access module owns CRUD/listing around access-right records.
- permission module consumes access repository for expansion/assignment behavior.
- access rights are authorization contract data, not only labels for UI.

## Route ownership

Access routes are registered under authorized group through:

`accessHttp.RegisterAccessRoutes(authorized.Group("", tenantMiddleware.OptionalOrganization()), accessModule.AccessController)`

Implications:

- route inherits admin-style authorization stack
- subgroup adds optional organization middleware again
- org context can matter when listing or scoping access data

## Contract semantics

Access-right records connect UI/admin intent to permission behavior.

Fields such as resource, action, endpoint, method, and names should be treated as security contract.

If these drift from route and Casbin semantics, permission UI may grant wrong or ineffective rights.

## Coupling map

Access-right changes can affect:

- permission assignment expansion
- role effective permissions
- user direct permissions
- admin UI option lists
- Casbin policy meaning if expansion maps to path/method semantics

Read with:

- `llm/cache/permission-system.md`
- `llm/cache/role-system.md`
- `llm/cache/casbin-permission-system.md`

## Known sharp edges

- renaming action/resource can silently change what future permission assignments mean.
- old policies may continue to exist after registry semantics change.
- optional organization route context must not be ignored.
- access-right records may be displayed like metadata but enforce like policy input.

## Change checklist

Before editing access-right code/data, prove:

1. Which permission expansion path consumes this access-right field?
2. Does existing policy data need migration or cleanup?
3. Does route/method/resource name match actual router behavior?
4. Does frontend permission UI consume this shape?
5. Does optional organization context affect output?

## Verification paths

- `internal/modules/access/test/*`
- `internal/modules/access/repository/access_repository_test.go`
- `internal/modules/access/usecase/access_usecase.go`
- `internal/modules/permission/usecase/access_right_assignment.go`
- `internal/modules/permission/test/*`
- `internal/router/router.go`

## Hard rules

- Treat access-right names/resources/actions as security contract.
- Do not rename semantics without tracing permission expansion and existing policies.
- Preserve authorized route behavior.
- Coordinate changes with permission and Casbin docs/code.
