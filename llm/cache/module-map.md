# Module Map

## Purpose

Durable map of backend modules, primary responsibilities, and high-signal dependencies so agents can route work to the correct owner layer quickly.

## Core runtime modules

| Module       | Location                        | Main responsibility                                           | High-signal dependencies                                         |
| ------------ | ------------------------------- | ------------------------------------------------------------- | ---------------------------------------------------------------- |
| auth         | `internal/modules/auth`         | login, register, session/token flows, SSO, ticket flow        | JWT manager, Redis/session, audit, worker, WS/SSE                |
| organization | `internal/modules/organization` | tenant lifecycle, members, invitations, cached org behavior   | user repo, Redis, enforcer, task distributor                     |
| user         | `internal/modules/user`         | registration, profile, avatar, user management                | auth, audit, webhook, storage, tx, Casbin interface              |
| permission   | `internal/modules/permission`   | policy CRUD, assignment, batch checks, access-right expansion | role repo, user repo, access repo, audit, transactional enforcer |
| access       | `internal/modules/access`       | access-right registry and endpoint/resource-action contract   | permission expansion consumers                                   |
| role         | `internal/modules/role`         | role CRUD, validation, cleanup orchestration                  | permission usecase, tx manager                                   |
| project      | `internal/modules/project`      | tenant-scoped project CRUD                                    | tenant routing, API-key scope, Casbin                            |
| api_key      | `internal/modules/api_key`      | API-key create/list/revoke/authenticate                       | organization repo, user repo, Redis, middleware                  |
| audit        | `internal/modules/audit`        | audit logs and outbox behavior                                | worker sync, organization visibility                             |
| webhook      | `internal/modules/webhook`      | webhook config, logs, async dispatch                          | worker distributor, tenant scope                                 |
| stats        | `internal/modules/stats`        | dashboard stats and metrics data                              | realtime broadcaster in app wiring                               |

## Routing rules

- If task is mostly route and contract: inspect route file plus `internal/router/router.go`
- If task is mostly business behavior: inspect target module `usecase` first
- If task is mostly persistence: inspect target module `repository` and `llm/conventions/database.md`
- If task crosses multiple modules, decide whether one module owns orchestration or whether it is truly cross-stack/cross-domain work

## Shared package hotspots

- `pkg/querybuilder`
- `pkg/tx`
- `pkg/storage`
- `pkg/tus`
- `pkg/ws`
- `pkg/sse`

These shared packages are often the true owner when changes look module-local but actually alter common infrastructure.

## Hard rules

- prefer module owner over scattered helper edits
- do not move business logic into controller layer
- if change crosses backend and frontend, update API contracts and proxy boundaries together
- if permission or tenant semantics move, trace router and middleware before patching module internals

## Evidence paths

- `internal/config/app.go`
- `internal/router/router.go`
- `internal/modules/*/module.go`
- relevant `llm/cache/*` domain files
