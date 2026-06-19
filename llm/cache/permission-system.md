# Permission System

## Purpose

Durable map for permission module behavior, access-right expansion, role/user permission orchestration, and transactional Casbin enforcement.

## Runtime truth

- `internal/modules/permission/module.go` wires enforcer, validator, logger, role repo, user repo, access-right repo, and audit module.
- `internal/modules/permission/delivery/http/permission_routes.go` registers permission routes under authorized/admin-style access.
- `internal/modules/permission/usecase/permission_usecase.go` owns permissions, role assignment, batch checks, and policy operations.
- `internal/modules/permission/usecase/transactional_enforcer.go` handles transactional Casbin access.

## Behavior surfaces

- permission policy CRUD/assignment
- role inheritance and access-right expansion
- batch permission checks
- transactional policy update/cleanup behavior

## Hard rules

- Policy writes sharing DB semantics should use transactional enforcer.
- Access-right changes must align with permission expansion behavior.
- Do not weaken security tests around Casbin failures.

## Tests and evidence paths

- `internal/modules/permission/test/*`
- `internal/modules/permission/usecase/*_test.go`
- `internal/modules/permission/usecase/transactional_enforcer.go`
