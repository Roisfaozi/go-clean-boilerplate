# Permission System

## Purpose

Durable map for permission-domain behavior:

- policy CRUD
- role and user permission assignment
- access-right expansion
- inheritance/effective permission calculation
- batch permission checks
- transactional Casbin writes

Use before changing `internal/modules/permission`, permission API routes, access-right assignment logic, or role/user policy semantics.

## Primary source of truth

1. `internal/modules/permission/delivery/http/permission_routes.go`
2. `internal/modules/permission/delivery/http/permission_controller.go`
3. `internal/modules/permission/usecase/permission_usecase.go`
4. `internal/modules/permission/usecase/access_right_assignment.go`
5. `internal/modules/permission/usecase/inheritance_tree.go`
6. `internal/modules/permission/usecase/transactional_enforcer.go`
7. `internal/router/router.go`
8. `llm/cache/access-right-system.md`
9. `llm/cache/casbin-permission-system.md`

## Runtime ownership

- permission module owns policy and assignment business behavior.
- transactional enforcer owns DB-transaction-aware Casbin adapter use.
- access-right assignment helper expands registry entries into permission rules.
- controller must bind/validate/respond, not own policy logic.

## Route ownership

Two route strata exist:

### Authenticated batch check

- registered with `permissionHttp.RegisterBatchCheckRoute(authenticated, ...)`
- app-facing permission check behavior
- not same as admin CRUD

### Authorized permission management

- registered with `permissionHttp.RegisterPermissionRoutes(authorized, ...)`
- admin-style policy CRUD and assignment behavior
- inherits `admin:manage` API-key scope, optional org, and Casbin middleware

This split must not collapse accidentally.

## Policy behavior surfaces

- add/update/remove permission policy
- grant/revoke permission to role
- grant/revoke role to user
- parent role inheritance
- role/user cleanup
- batch enforce checks
- access-right expansion from registry

## Transaction semantics

When permission writes are tied to DB state:

- use `IEnforcer.WithContext(ctx)` path
- preserve transaction-bound enforcer behavior
- avoid raw global enforcer shortcuts inside transactional flows

This matters for role/user flows that must not commit DB state without matching policy state.

## Cross-system coupling

Permission changes can affect:

- access-right registry meaning
- role CRUD and role deletion cleanup
- user role assignment
- tenant/domain authorization
- API-key/admin route behavior
- frontend permission UI and available action lists

## Known sharp edges

- batch-check route is authenticated, not admin-only.
- access-right expansion can multiply one UI action into many policies.
- policy write failure must not be swallowed.
- role inheritance changes can alter effective permission far from edited code.
- invalid inputs and Casbin failures need negative tests.

## Change checklist

Before editing permission code, prove:

1. Is target batch check or admin policy management?
2. Does change write policy, grouping policy, or both?
3. Does access-right expansion still match registry semantics?
4. Does change need transactional enforcer context?
5. Does route/domain behavior depend on tenant or optional org context?
6. Do frontend permission screens need new action/resource labels?

## Verification paths

- `internal/modules/permission/test/*`
- `internal/modules/permission/usecase/permission_usecase.go`
- `internal/modules/permission/usecase/transactional_enforcer.go`
- `internal/modules/permission/usecase/access_right_assignment.go`
- `internal/modules/permission/usecase/inheritance_tree.go`
- `internal/router/router.go`

## Hard rules

- Do not merge batch-check semantics with admin CRUD semantics.
- Do not bypass transactional enforcer when DB and policy writes must stay atomic.
- Do not weaken negative-path tests around Casbin errors.
- Keep access-right expansion aligned with registry data.
