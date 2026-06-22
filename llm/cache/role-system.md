# Role System

## Purpose

Durable map for role-domain behavior:

- role CRUD
- validation and model conversion
- dynamic role search/list
- role-policy cleanup
- permission-usecase orchestration
- admin route protection

Use before changing `internal/modules/role`, role API contracts, role validation, role deletion, or permission cleanup tied to role records.

## Primary source of truth

1. `internal/modules/role/delivery/http/role_routes.go`
2. `internal/modules/role/delivery/http/role_controller.go`
3. `internal/modules/role/usecase/role_usecase.go`
4. `internal/modules/role/repository/role_repository.go`
5. `internal/modules/role/module.go`
6. `internal/modules/permission/usecase/*`
7. `internal/router/router.go`

## Runtime ownership

- `NewRoleModule` wires DB, logger, validator, transaction manager, role repository, and permission usecase.
- role controller owns request bind/validate/response only.
- role usecase owns business behavior and permission cleanup orchestration.
- role repository owns DB reads/writes.

## Route ownership

Role routes are registered under `authorized` route group.

That means role routes inherit:

1. API-key authentication
2. JWT/session validation
3. explicit `admin:manage` API-key scope for API-key actors
4. user status middleware
5. optional organization context
6. Casbin middleware

Do not treat role endpoints as self-service or tenant CRUD by default.

## Behavior surfaces

- create role
- get all roles
- update role
- delete role
- dynamic filter/search path in controller tests
- validation through role model and custom validation
- deletion or update cleanup through permission usecase

## Coupling to permission and Casbin

Role changes can affect effective authorization without editing Casbin middleware.

Important coupling:

- permission usecase can own policy cleanup for role deletion or role changes
- access-right expansion can change what a role means
- role naming/normalization can affect grouping policies
- transaction manager matters when DB role state and permission state must stay consistent

## Known sharp edges

- role deletion is not just row delete when policies reference role.
- role rename/update can invalidate expectations in permission assignments.
- dynamic role search must obey querybuilder safety rules.
- optional organization on authorized group can affect domain-sensitive behavior.

## Change checklist

Before editing role code, prove:

1. Does change alter role name, slug, validation, or uniqueness semantics?
2. Does change need permission cleanup or policy update?
3. Does DB state and policy state need transaction consistency?
4. Does route remain admin-only through authorized group?
5. Does dynamic search expose new fields?

## Verification paths

- `internal/modules/role/usecase/*_test.go`
- `internal/modules/role/repository/role_repository_test.go`
- `internal/modules/role/delivery/http/role_controller_test.go`
- `internal/modules/role/model/*_test.go`
- `internal/modules/permission/test/*` when cleanup changes
- `internal/router/router.go`

## Hard rules

- Do not bypass permission usecase integration for role-policy cleanup.
- Do not move role business logic into controller.
- Do not weaken authorized route assumptions.
- Do not change role identity semantics without tracing policies and assignments.
