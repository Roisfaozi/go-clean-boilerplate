# Permission System

## Purpose

Durable map for permission policy behavior, role and user assignment, access-right expansion, batch permission checks, and transactional Casbin enforcement.

## Runtime truth

- Module root: `internal/modules/permission/`
- Wiring entry: `internal/modules/permission/module.go`
- Routes: `internal/modules/permission/delivery/http/permission_routes.go`
- Controller: `internal/modules/permission/delivery/http/permission_controller.go`
- Usecase: `internal/modules/permission/usecase/permission_usecase.go`
- Transaction boundary helper: `internal/modules/permission/usecase/transactional_enforcer.go`
- Access-right expansion logic: `internal/modules/permission/usecase/access_right_assignment.go`
- Inheritance logic: `internal/modules/permission/usecase/inheritance_tree.go`

## Route ownership

- Batch permission check route is registered under `authenticated` routes through `permissionHttp.RegisterBatchCheckRoute(...)`.
- Admin-style permission CRUD and assignment routes are registered under `authorized` routes through `permissionHttp.RegisterPermissionRoutes(...)`.
- This split is important: self-service or app-facing checks are not the same as admin policy management.

## Behavior surfaces

- permission policy CRUD
- role assignment and cleanup
- user assignment and cleanup
- access-right expansion
- inheritance-based effective permission calculation
- batch permission checks
- transactional Casbin write behavior

## Coupling to other systems

- depends on access-right registry semantics from `internal/modules/access`
- interacts with role and user repositories
- interacts with Casbin policy state and startup enforcement guarantees
- can emit audit side effects via module wiring

## Hard rules

- Policy writes sharing DB semantics should go through transactional enforcer patterns.
- Access-right changes must stay aligned with permission expansion behavior.
- Do not weaken negative-path security tests around Casbin failures or invalid inputs.
- Do not merge batch-check semantics and admin policy semantics by accident.

## Verification and evidence paths

- `internal/modules/permission/test/*`
- `internal/modules/permission/usecase/permission_usecase.go`
- `internal/modules/permission/usecase/transactional_enforcer.go`
- `internal/modules/permission/usecase/access_right_assignment.go`
- `internal/router/router.go`
