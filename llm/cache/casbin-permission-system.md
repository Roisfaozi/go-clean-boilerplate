# Casbin Permission System

## Purpose

Durable map for Casbin policy/enforcer, permission module, route authorization, and transactional policy behavior.

## Runtime truth

- Casbin enforcer is initialized and production-guarded in `internal/config/app.go`.
- `internal/config/app.go` fails startup in production if Casbin is disabled or loaded with zero policies.
- `internal/modules/permission` owns policy operations, role/user permissions, access-right expansion, inheritance, and batch checks.
- `internal/modules/permission/usecase/transactional_enforcer.go` is the repo-specific transactional enforcer boundary.
- `internal/middleware/casbin_middleware.go` is the HTTP authorization middleware.

## Route authorization layering

- `tenantAuthorized` routes require tenant context before Casbin middleware.
- `authorized` routes require API-key admin scope, optional org context, and Casbin middleware.
- Batch permission check route is registered under authenticated routes.

## Hard rules

- Casbin policy writes tied to DB state should go through permission/Casbin abstractions and transactional enforcer patterns.
- Do not move Casbin checks from router/middleware to frontend-only checks.
- Do not weaken matcher/path behavior casually; route object/method/domain semantics are access-control sensitive.
- Production must not fail open with disabled or empty-policy Casbin.

## Tests and evidence paths

- `internal/modules/permission/test/*`
- `internal/modules/auth/repository/casbin_adapter_test.go`
- `internal/middleware/casbin_middleware.go`
- `internal/router/router.go`
- `internal/config/app.go`
