# Module Map

## Purpose

Durable routing map for backend module ownership in this repo.

Use this file to answer:

- which module owns business rule
- which module only exposes route/controller shell
- which dependency edges make a change cross-domain
- when change that looks local is really shared-infra work

This file is for fast ownership decisions before deep code reads.

## Primary source of truth

1. `internal/config/app.go`
2. `internal/router/router.go`
3. `internal/modules/*/module.go`
4. relevant `llm/cache/*` domain files

## Composition-root facts

`internal/config/app.go` is real dependency graph.

It wires:

- modules
- middleware
- storage
- Redis
- Casbin
- task distributor/processor
- websocket and sse managers
- TUS registry and handler

If module ownership is unclear, inspect app wiring before guessing from package names.

## Module ownership table

| Module         | Entrypoint                                | Primary responsibility                                                                        | High-signal dependencies                                                 | Route strata                                       |
| -------------- | ----------------------------------------- | --------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------ | -------------------------------------------------- |
| `auth`         | `internal/modules/auth/module.go`         | login, register-adjacent auth, refresh/logout, session validation, SSO, ticket issuance       | JWT, Redis token repo, audit, worker, organization repo, ws, sse, Casbin | public + authenticated                             |
| `organization` | `internal/modules/organization/module.go` | organization lifecycle, membership, invitations, tenant reader/cache, restore/admin org flows | Redis, task distributor, user repo, enforcer, presence manager           | public + authenticated + tenant + admin            |
| `user`         | `internal/modules/user/module.go`         | registration, self profile, avatar, admin user management                                     | tx, enforcer, audit, auth, webhook, storage                              | public + authenticated + admin                     |
| `permission`   | `internal/modules/permission/module.go`   | permission CRUD, role/user assignment, batch permission checks, policy expansion              | enforcer, role repo, user repo, access repo, audit                       | authenticated + admin                              |
| `access`       | `internal/modules/access/module.go`       | access-right registry and endpoint/resource-action catalog                                    | DB, validation, permission consumers downstream                          | admin                                              |
| `role`         | `internal/modules/role/module.go`         | role CRUD, role validation, role-policy cleanup orchestration                                 | tx manager, permission usecase, role repo                                | admin                                              |
| `project`      | `internal/modules/project/module.go`      | tenant-scoped project CRUD                                                                    | tenant route group, API-key scope, Casbin                                | tenant-authorized                                  |
| `api_key`      | `internal/modules/api_key/module.go`      | API-key create/list/revoke/authenticate and scope identity                                    | user repo, Redis, middleware coupling                                    | authenticated + tenant/admin enforcement consumers |
| `audit`        | `internal/modules/audit/module.go`        | audit logging, outbox, list/search, cleanup                                                   | ws manager, task distributor, querybuilder                               | admin + worker side effects                        |
| `stats`        | `internal/modules/stats/module.go`        | dashboard summaries and system insights                                                       | DB, realtime broadcaster in app wiring                                   | authenticated                                      |
| `webhook`      | `internal/modules/webhook/module.go`      | webhook CRUD, logging, trigger dispatch                                                       | worker task distributor, validation                                      | tenant-authorized                                  |

## Route-group map

### Public-owned module surfaces

- `auth`
- `user` registration
- `organization` public bootstrap/acceptance paths

### Authenticated-owned module surfaces

- `auth` session endpoints
- `user` self profile
- `organization` list/create for current member
- `permission` batch check
- `api_key` CRUD
- `stats`

### Tenant-authorized module surfaces

- `organization` tenant member/invite/presence style flows
- `project`
- `webhook`

### Admin/authorized module surfaces

- `organization` restore/admin
- `permission`
- `access`
- `role`
- `user` admin management
- `audit`

Route strata matter because same module may appear in multiple access layers with different guarantees.

## Dependency patterns that usually mean cross-domain work

Treat change as cross-domain, not module-local, when module touches any of these:

- auth + Redis session state
- tenant/org context
- Casbin enforcer or permission usecase
- worker task distributor
- TUS/storage/avatar flow
- `pkg/querybuilder`
- frontend API contracts or proxy assumptions

Examples:

- changing `user` registration likely touches auth, audit, webhook, Casbin
- changing `project` route semantics likely touches tenant middleware, API-key scope, Casbin, frontend contract
- changing `organization` membership can affect presence, cache invalidation, tenant reader, and route auth

## Shared package hotspots

These often own real behavior even when issue appears module-local:

- `pkg/querybuilder` — dynamic filter/sort security
- `pkg/tx` — transaction propagation
- `pkg/storage` — file/object storage abstraction
- `pkg/tus` — resumable upload path and hook dispatch
- `pkg/ws` — websocket/ticket/presence behavior
- `pkg/sse` — SSE fanout behavior
- `pkg/database` — org context propagation

## Ownership heuristics

Use these fast rules:

- if route protection changed, start in `internal/router/router.go` before module code
- if DB + Casbin must commit together, owner is usecase + transactional enforcer, not repository only
- if async side effect exists, inspect module plus worker/task distributor
- if frontend breaks after backend patch, inspect `llm/cache/api-contracts.md` and frontend proxy files before blaming frontend alone
- if field-level search/list behavior changed, check `pkg/querybuilder` before editing repositories broadly

## Known sharp edges

- module names hide real coupling; `user` and `organization` are especially cross-cutting.
- `stats` module looks isolated but realtime metrics broadcaster is wired outside module in app root.
- `api_key` behavior is split between module usecase and router/middleware layering.
- `permission` behavior is inseparable from `role`, `access`, and tenant domain semantics.

## Hard rules

- Prefer module owner over scattered helper edits.
- Do not move business logic into controllers.
- If permission or tenant semantics move, inspect router and middleware before patching module internals.
- If change crosses backend and frontend, update API contract thinking together.
