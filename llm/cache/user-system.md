# User System

## Purpose

Durable map for user domain behavior: profile, status, avatar, dynamic list/search, audit/webhook side effects, and dependency wiring.

## Runtime truth

- Module root: `internal/modules/user/module.go`.
- Controller: `internal/modules/user/delivery/http/user_controller.go`.
- Routes: `internal/modules/user/delivery/http/user_routes.go`.
- Usecase: `internal/modules/user/usecase/user_usecase.go`.
- Repository: `internal/modules/user/repository/user_repository.go`.
- Avatar hook: `internal/modules/user/usecase/avatar_hook.go`.
- Module wiring in `internal/config/app.go` passes DB, logger, validator, transaction manager, Casbin enforcer, audit module, auth module, webhook module, and storage provider.

## Dependencies

`NewUserModule` wires:

- `tx.WithTransactionManager`
- permission/Casbin enforcer interface
- audit usecase
- auth usecase
- webhook usecase
- storage provider

## Route ownership

User routes are registered in multiple route strata through `internal/router/router.go`:

- public user routes through `userHttp.RegisterPublicRoutes`
- authenticated user routes through `userHttp.RegisterAuthenticatedRoutes`
- authorized/admin-style user routes through `userHttp.RegisterAuthorizedRoutes`

## Behavior surfaces

- user registration route exists in public/user tests and route registration.
- profile/self-update and avatar update behavior are tested in controller/usecase tests.
- admin-style user list/detail/delete/status/dynamic search routes exist in tests.
- dynamic search/list behavior can intersect `pkg/querybuilder` security rules.
- avatar upload is tied to storage/TUS completion hook behavior.

## Hard rules

- User changes can affect audit, auth/session, webhook, storage, and Casbin; classify those boundaries before patching.
- Do not leak business logic into controller; controller binds/validates/responds.
- Do not weaken dynamic search/querybuilder restrictions for user fields.
- Avatar/storage changes must preserve context propagation and storage provider abstraction.

## Tests and evidence paths

- `internal/modules/user/test/*`
- `internal/modules/user/delivery/http/user_controller_test.go`
- `internal/modules/user/repository/user_repository_test.go`
- `internal/modules/user/test/avatar_hook_test.go`
