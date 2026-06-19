# User System

## Purpose

Durable map for user domain behavior: registration, profile, status, avatar, admin list/search, and side effects into auth, audit, webhook, storage, and permission boundaries.

## Runtime truth

- Module root: `internal/modules/user/`
- Wiring entry: `internal/modules/user/module.go`
- Controller: `internal/modules/user/delivery/http/user_controller.go`
- Routes: `internal/modules/user/delivery/http/user_routes.go`
- Usecase: `internal/modules/user/usecase/user_usecase.go`
- Repository: `internal/modules/user/repository/user_repository.go`
- Avatar hook: `internal/modules/user/usecase/avatar_hook.go`
- Module wiring in `internal/config/app.go` passes DB, logger, validator, transaction manager, Casbin enforcer interface, audit module, auth module, webhook module, and storage provider.

## Route ownership

User routes are registered in multiple strata through `internal/router/router.go`:

- public routes through `userHttp.RegisterPublicRoutes`
- authenticated routes through `userHttp.RegisterAuthenticatedRoutes`
- authorized/admin-style routes through `userHttp.RegisterAuthorizedRoutes`

This means user domain is not one single access shape. Registration, self-service, and admin management must be treated separately.

## Behavior surfaces

- user registration
- current-user or self-service profile flows
- avatar update behavior
- admin-style user list, detail, delete, restore/status, and dynamic search behavior
- user-related side effects that can touch audit, auth, webhook, or storage

## Dependency and coupling map

`NewUserModule` wires:

- `tx.WithTransactionManager`
- permission or Casbin enforcer interface
- audit usecase
- auth usecase
- webhook usecase
- storage provider

This makes user changes sensitive even when the edited file looks local.

## Query and avatar specifics

- dynamic search/list behavior can intersect `pkg/querybuilder` security rules
- avatar upload is tied to storage and TUS completion-hook behavior
- avatar/storage changes must preserve context propagation and storage provider abstraction

## Hard rules

- User changes can affect audit, auth/session, webhook, storage, and Casbin; classify those boundaries before patching.
- Do not leak business logic into controller; controller binds, validates, and responds.
- Do not weaken dynamic search or querybuilder restrictions for user fields.
- Do not treat avatar flow as isolated UI-only behavior when storage and upload hooks are involved.

## Verification and evidence paths

- `internal/modules/user/test/*`
- `internal/modules/user/delivery/http/user_controller_test.go`
- `internal/modules/user/repository/user_repository_test.go`
- `internal/modules/user/test/avatar_hook_test.go`
- `internal/modules/user/usecase/user_usecase.go`
- `internal/router/router.go`
