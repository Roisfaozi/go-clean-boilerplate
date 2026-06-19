# Casbin Permission System

## Purpose

Durable map for Casbin enforcer lifecycle, permission module ownership, route authorization layering, and transactional policy behavior.

## Runtime truth

- Casbin enforcer is initialized in `internal/config/app.go`.
- Startup fails in production if Casbin is disabled or loaded with zero policies.
- `internal/modules/permission` owns policy operations, role and user permissions, access-right expansion, inheritance, and batch checks.
- `internal/modules/permission/usecase/transactional_enforcer.go` is the repo-specific transactional enforcer boundary.
- `internal/middleware/casbin_middleware.go` is the HTTP authorization middleware.

## Route authorization layering

- `tenantAuthorized` routes require tenant context before Casbin middleware.
- `authorized` routes require explicit admin API-key scope, optional org context, and Casbin middleware.
- Batch permission check route is registered under authenticated routes, not admin-only routes.

## Enforcement inputs

Effective authorization can depend on:

- authenticated user or API-key actor identity
- organization or tenant domain context
- request path and method
- expanded access-right semantics
- role and user assignments

## Policy ownership

- direct policy CRUD, assignment, expansion, and cleanup belong to permission module abstractions
- policy writes tied to DB state should use transactional patterns
- route protection decisions still belong in router plus middleware layering, not frontend-only checks

## Coupling to other systems

- access-right registry changes can affect effective Casbin behavior
- role and permission flows are tightly coupled
- tenant middleware ordering matters because domain context feeds authorization
- startup safety guarantees in `internal/config/app.go` prevent fail-open production behavior

## Hard rules

- Casbin policy writes tied to DB state should go through permission and transactional enforcer patterns.
- Do not move Casbin checks from router or middleware into frontend-only checks.
- Do not weaken matcher, path, method, or domain semantics casually.
- Production must not fail open with disabled or empty-policy Casbin.

## Verification and evidence paths

- `internal/modules/permission/test/*`
- `internal/modules/auth/repository/casbin_adapter_test.go`
- `internal/modules/permission/usecase/transactional_enforcer.go`
- `internal/middleware/casbin_middleware.go`
- `internal/router/router.go`
- `internal/config/app.go`
