# User System

## Purpose

Durable map for user-domain behavior in this repo:

- registration and default-role side effects
- self-service profile endpoints
- admin-style user management routes
- avatar upload and post-upload hook behavior
- dynamic user listing/search through querybuilder
- coupling into auth, audit, webhook, storage, and permission systems

Use this file before changing `internal/modules/user`, avatar flow, user list/search behavior, status changes, delete behavior, or any user-facing API contract.

## Primary source of truth

1. `internal/modules/user/delivery/http/user_routes.go`
2. `internal/modules/user/delivery/http/user_controller.go`
3. `internal/modules/user/usecase/user_usecase.go`
4. `internal/modules/user/repository/user_repository.go`
5. `internal/modules/user/usecase/avatar_hook.go`
6. `internal/config/app.go`
7. `llm/cache/querybuilder-security.md`
8. `llm/cache/tus-upload-system.md`
9. `llm/cache/authentication-system.md`

## Runtime ownership

### Module wiring

`internal/modules/user/module.go` proves user module is not standalone CRUD.

`NewUserModule` wires:

- transaction manager
- Casbin/permission enforcer interface
- audit usecase
- auth usecase
- webhook usecase
- storage provider

Implication:

- user change can affect role assignment, audit, async webhook, avatar storage, and auth lifecycle
- controller-only patch is often not enough to understand impact

### Route ownership

User routes live in three access strata.

#### Public

`RegisterPublicRoutes`:

- `POST /users/register`

#### Authenticated self-service

`RegisterAuthenticatedRoutes`:

- `GET /users/me`
- `PUT /users/me`
- `PATCH /users/me/avatar`

These are human-session routes, not admin surfaces.

#### Authorized/admin-style

`RegisterAuthorizedRoutes`:

- `GET /users`
- `POST /users/search`
- `GET /users/:id`
- `PATCH /users/:id/status`
- `DELETE /users/:id`

These run under `authorized` group in `internal/router/router.go`, so they inherit:

1. API-key authenticate
2. token validate
3. `admin:manage` API-key scope
4. user-status middleware
5. optional organization
6. Casbin middleware

Meaning:

- admin user routes are protected by more than one layer
- changing one layer does not replace others

## Core behavior surfaces

### Registration

`internal/modules/user/usecase/user_usecase.go` proves registration does more than insert row:

- sanitizes name and username
- checks username and email uniqueness
- hashes password with bcrypt
- creates UUIDv7 user ID
- creates user in transaction
- assigns default Casbin grouping policy `role:user` in `global` domain when enforcer present
- writes audit log in same transaction path
- compensates role assignment on audit failure
- triggers `user.created` webhook out of transaction

This means registration touches user data, default authorization, audit, and async integrations.

### Self profile

Self-service routes live under authenticated group and should be treated separately from admin edits.

If changing `/users/me` or avatar update flow, audit:

- token/session semantics
- returned profile shape
- storage side effects
- frontend profile surfaces

### Admin list and search

User list/search endpoints are especially sensitive because they combine privileged data access with dynamic query behavior.

- `GET /users` supports admin list behavior
- `POST /users/search` uses dynamic filter path
- any query surface widening must be reviewed with `llm/cache/querybuilder-security.md`

### Status and delete

- `PATCH /users/:id/status` is privileged status mutation
- `DELETE /users/:id` is privileged destructive action
- repository also contains hard-delete cleanup path for soft-deleted users

Do not treat these as cosmetic admin actions; auth/session and audit expectations may be involved.

## Avatar behavior

User avatar path has two related entrypoints:

### Direct authenticated avatar update

- `PATCH /users/me/avatar`
- controller/usecase path handles user-facing avatar mutation

### TUS completion hook avatar update

- `pkg/tus` completion dispatch chooses `avatar` hook
- `internal/config/app.go` registers `AvatarHook`
- `internal/modules/user/test/avatar_hook_test.go` proves `authenticated_user_id` must take precedence over client `user_id`

Implication:

- avatar logic is split across direct user API and upload-completion path
- changing one path without other can drift behavior

## Cross-system coupling

User changes can silently break:

- `auth` because registration and current-user flows depend on auth/session truth
- `permission` / Casbin because default role assignment and admin access are linked
- `audit` because create/status/delete/avatar flows may emit logs
- `webhook` because user lifecycle emits external events
- `storage` and `tus` because avatar URLs and upload completion mutate user data
- `querybuilder` because admin search/list surface exposes model-backed fields

Read with:

- `llm/cache/authentication-system.md`
- `llm/cache/casbin-permission-system.md`
- `llm/cache/querybuilder-security.md`
- `llm/cache/tus-upload-system.md`

## Known sharp edges

- registration transaction includes role assignment and audit semantics; partial rewrite can create inconsistent state.
- user list/search is security-sensitive even if only response formatting seems to change.
- avatar flow has both direct API path and async TUS hook path.
- admin routes sit behind API-key scope plus Casbin; do not simplify them into one check mentally.
- webhook trigger is out-of-transaction; side-effect timing matters.

## Change checklist

Before editing user-domain code, prove these answers:

1. Is target flow public registration, self-service profile, or admin management?
2. Does change affect default role assignment, audit logging, or webhook trigger?
3. Does list/search behavior expose new filterable or sortable fields?
4. Is avatar change on direct API path, TUS hook path, or both?
5. Do frontend apps consume this user response shape?
6. Does change alter auth/session assumptions for `/users/me` or registration?

## Verification paths

Narrow first:

- `internal/modules/user/test/*`
- `internal/modules/user/test/avatar_hook_test.go`
- `internal/modules/user/repository/user_repository_test.go`
- `internal/modules/user/delivery/http/user_controller_test.go`
- `internal/modules/user/usecase/user_usecase.go`

Broader when auth/avatar/list behavior changed:

- target frontend typecheck/build
- integration auth/profile/avatar flow if touched

## Hard rules

- Do not move business rules into controller layer.
- Do not weaken querybuilder restrictions on user fields.
- Do not treat avatar flow as UI-only; it spans storage and async hook boundaries.
- Do not forget registration side effects beyond row creation.
- Do not bypass admin route layering when changing user management endpoints.
