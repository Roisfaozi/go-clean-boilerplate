# Module Map

## Backend modules

| Module       | Location                        | Main responsibility                                   | Key dependencies                                                   |
| ------------ | ------------------------------- | ----------------------------------------------------- | ------------------------------------------------------------------ |
| access       | `internal/modules/access`       | access-right and endpoint registry                    | GORM, validator, logger                                            |
| api_key      | `internal/modules/api_key`      | API key create/list/revoke/authenticate               | user repo, org repo, Redis                                         |
| audit        | `internal/modules/audit`        | audit log, outbox, broadcast                          | GORM, WebSocket, worker                                            |
| auth         | `internal/modules/auth`         | register/login/session/reset/SSO/ticket               | JWT, Redis token repo, user repo, org repo, Casbin adapter, worker |
| organization | `internal/modules/organization` | org lifecycle, membership, invitations, tenant reader | Redis, worker, user repo, enforcer, presence reader                |
| permission   | `internal/modules/permission`   | Casbin policy operations and access-right expansion   | enforcer, role/user/access repos, audit                            |
| project      | `internal/modules/project`      | tenant-scoped project CRUD                            | project repository                                                 |
| role         | `internal/modules/role`         | role CRUD and cleanup                                 | role repo, permission usecase, transaction manager                 |
| stats        | `internal/modules/stats`        | dashboard summary/activity/insights                   | GORM                                                               |
| user         | `internal/modules/user`         | user CRUD/profile/status/avatar                       | tx, enforcer, audit, auth, webhook, storage                        |
| webhook      | `internal/modules/webhook`      | webhook configs/logs/dispatch                         | repository, worker, validator                                      |

## Module composition patterns

- `auth` is the heaviest composition point: it bridges token storage, JWT, SSO, org bootstrap, worker queueing, and event publishing.
- `organization` owns tenant context and member lifecycle; it is the main boundary for org/invitation/membership logic.
- `permission` is the policy transformation layer between roles/access-rights and Casbin enforcement.
- `user` is the shared domain center for profile, avatar, status, and admin/user-management flows.
- `audit` and `webhook` are side-effect modules that often follow writes in other modules.
- `api_key` is both auth-adjacent and tenant-adjacent; it should be checked when changing protected routes.

## Frontend modules

`apps/web` module surfaces:

- auth pages/actions
- dashboard pages
- API proxy route handlers
- shared UI/components/hooks/stores

`apps/web` quality notes:

- App Router route groups split auth, dashboard, and API handler concerns.
- It is the most integration-heavy frontend because it mixes server components, server actions, and backend proxying.

`apps/client` module surfaces:

- feature folders for users/roles/organizations/projects/permissions/resources/endpoints/audit logs
- route registry in `app/routes.ts`
- shared components/lib/stores/hooks

`apps/client` quality notes:

- Route registry is the main navigation contract.
- Feature folders are domain-first, which makes them easy to map to backend modules.
- API proxy route is a transport layer, not a business layer.

## Cross-module rules

- Backend module constructors should receive only required primitive/config/dependency values, not full app config.
- Cross-module dependencies are wired in `internal/config/app.go`, not inside controllers.
- Casbin policy writes should go through permission/Casbin abstractions, especially if transaction-bound.
- Tenant-aware behavior should resolve organization context through middleware/usecase/repository paths, not ad hoc query params only.
- Shared workspace packages should be preferred over duplicating UI/api helper code across apps.
- If a change crosses backend and frontend, update module map and API contracts together so routing and transport stay aligned.
