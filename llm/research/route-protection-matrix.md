# Route Protection Matrix

## Scope

Matrix ini mendokumentasikan route groups, middleware semantics, dan gap coverage berdasarkan live code.

## Evidence Paths

- `internal/router/router.go`
- `internal/modules/access/delivery/http/access_routes.go`
- `internal/modules/api_key/delivery/http/api_key_routes.go`
- `internal/modules/audit/delivery/http/audit_routes.go`
- `internal/modules/auth/delivery/http/auth_routes.go`
- `internal/modules/organization/delivery/http/organization_routes.go`
- `internal/modules/permission/delivery/http/permission_routes.go`
- `internal/modules/role/delivery/http/role_routes.go`
- `internal/modules/user/delivery/http/user_routes.go`
- `internal/modules/webhook/delivery/http/webhook_routes.go`
- `tests/integration/*`
- `tests/e2e/*`

## Group Semantics

### `public`

- Base middlewares: public limiter only when enabled.
- No JWT required.
- No API key.
- No tenant context.
- No Casbin.

### `authenticated`

- `apiKeyMiddleware.Authenticate()`
- `authMiddleware.ValidateToken()`
- `apiKeyMiddleware.RequireScopeAuto()`
- `apiKeyMiddleware.RequireUserSession()`
- `UserStatusMiddleware`
- Auth limiter when enabled.

Truth:

- JWT/session required.
- API key machinery present, but API-key-only access blocked by `RequireUserSession()`.
- No tenant required by group itself.
- No Casbin at group level.

### `tenantAuthorized`

- `apiKeyMiddleware.Authenticate()`
- `authMiddleware.ValidateToken()`
- `apiKeyMiddleware.RequireScopeAuto()`
- `UserStatusMiddleware`
- `tenantMiddleware.RequireOrganization()`
- `casbinMiddleware`
- Auth limiter when enabled.

Truth:

- JWT or API key auth path allowed.
- Tenant required.
- Casbin domain should be org-scoped.
- Route-level explicit scopes may narrow further.

### `authorized`

- `apiKeyMiddleware.Authenticate()`
- `authMiddleware.ValidateToken()`
- `apiKeyMiddleware.RequireScopes("admin:manage")`
- `UserStatusMiddleware`
- `tenantMiddleware.OptionalOrganization()`
- `casbinMiddleware`
- Auth limiter when enabled.

Truth:

- JWT or API key auth path allowed.
- Admin scope required.
- Casbin domain can be `global` or resolved org.

### `upload`

- `authMiddleware.ValidateToken()`
- `UserStatusMiddleware`

Truth:

- JWT/session required.
- No API key path.
- No tenant middleware.
- No Casbin at router layer.

### `realtime`

- SSE `/events`: `authMiddleware.ValidateToken()`
- WS `/ws`: `authMiddleware.ValidateWebSocketToken()`

Truth:

- SSE uses JWT/session path.
- WS uses ticket path.

## Representative Route Matrix

| Route                                       | Method | Group                            | JWT Session        | API Key                               | User Session Required | Org Required             | Casbin                 | Extra Scope                     | Coverage Signal                                         |
| ------------------------------------------- | ------ | -------------------------------- | ------------------ | ------------------------------------- | --------------------- | ------------------------ | ---------------------- | ------------------------------- | ------------------------------------------------------- |
| `/api/v1/auth/login`                        | POST   | public                           | no                 | no                                    | no                    | no                       | no                     | none                            | auth integration/e2e present                            |
| `/api/v1/auth/refresh`                      | POST   | public                           | refresh token flow | no                                    | no                    | no                       | no                     | none                            | auth integration/e2e present                            |
| `/api/v1/auth/register`                     | POST   | public                           | no                 | no                                    | no                    | no                       | no                     | none                            | auth integration present                                |
| `/api/v1/auth/sso/:provider`                | GET    | public                           | no                 | no                                    | no                    | no                       | no                     | none                            | SSO unit coverage partial, route E2E unclear            |
| `/api/v1/users/register`                    | POST   | public                           | no                 | no                                    | no                    | no                       | no                     | none                            | user integration likely, exact route needs confirmation |
| `/api/v1/organizations/invitations/accept`  | POST   | public                           | no                 | no                                    | no                    | no                       | no                     | none                            | organization integration present                        |
| `/api/v1/events`                            | GET    | realtime                         | yes                | no                                    | yes                   | no                       | no                     | none                            | SSE E2E present                                         |
| `/api/v1/ws`                                | GET    | realtime                         | ticket             | no                                    | ticket-backed         | org optional from ticket | no                     | none                            | WebSocket E2E present                                   |
| `/api/v1/health`                            | GET    | infra                            | no                 | no                                    | no                    | no                       | no                     | none                            | covered indirectly only                                 |
| `/api/v1/auth/logout`                       | POST   | authenticated                    | yes                | no effective                          | yes                   | no                       | no                     | auto scope                      | auth integration/e2e present                            |
| `/api/v1/auth/ticket`                       | POST   | authenticated                    | yes                | no effective                          | yes                   | no                       | no                     | auto scope                      | auth middleware/unit coverage; route E2E unclear        |
| `/api/v1/auth/me`                           | GET    | authenticated                    | yes                | no effective                          | yes                   | no                       | no                     | auto scope                      | auth e2e likely                                         |
| `/api/v1/stats/summary`                     | GET    | authenticated                    | yes                | no effective                          | yes                   | no                       | no                     | auto scope                      | stats integration/e2e present                           |
| `/api/v1/users/me`                          | GET    | authenticated                    | yes                | no effective                          | yes                   | no                       | no                     | auto scope                      | user e2e present                                        |
| `/api/v1/users/me`                          | PUT    | authenticated                    | yes                | no effective                          | yes                   | no                       | no                     | auto scope                      | user integration/e2e partial                            |
| `/api/v1/users/me/avatar`                   | PATCH  | authenticated                    | yes                | no effective                          | yes                   | no                       | no                     | auto scope                      | avatar/upload path needs stronger cross-check           |
| `/api/v1/organizations`                     | POST   | authenticated                    | yes                | no effective                          | yes                   | no                       | no                     | auto scope                      | organization integration/e2e present                    |
| `/api/v1/organizations/me`                  | GET    | authenticated                    | yes                | no effective                          | yes                   | no                       | no                     | auto scope                      | organization integration present                        |
| `/api/v1/permissions/check-batch`           | POST   | authenticated                    | yes                | no effective                          | yes                   | no                       | no                     | auto scope                      | permission batch integration present                    |
| `/api/v1/api-keys`                          | POST   | authenticated + nested tenant    | yes                | no effective                          | yes                   | yes                      | no direct group Casbin | nested org required             | api key integration present                             |
| `/api/v1/organizations/:id`                 | GET    | tenantAuthorized                 | yes or api key     | yes                                   | no                    | yes                      | yes                    | `org:view`/`org:manage`         | organization e2e/integration present                    |
| `/api/v1/organizations/slug/:slug`          | GET    | tenantAuthorized                 | yes or api key     | yes                                   | no                    | yes                      | yes                    | `org:view`/`org:manage`         | route coverage needs confirmation                       |
| `/api/v1/organizations/:id`                 | PUT    | tenantAuthorized                 | yes or api key     | yes                                   | no                    | yes                      | yes                    | `org:manage`                    | organization e2e partial                                |
| `/api/v1/organizations/:id`                 | DELETE | tenantAuthorized                 | yes                | api key denied by nested user-session | yes for nested route  | yes                      | yes                    | none + `RequireUserSession()`   | org delete tests present                                |
| `/api/v1/organizations/:id/members/invite`  | POST   | tenantAuthorized                 | yes or api key     | yes                                   | no                    | yes                      | yes                    | `member:manage`                 | organization integration present                        |
| `/api/v1/organizations/:id/members`         | GET    | tenantAuthorized                 | yes or api key     | yes                                   | no                    | yes                      | yes                    | `member:manage`                 | presence/member tests partial                           |
| `/api/v1/organizations/:id/members/:userId` | PATCH  | tenantAuthorized                 | yes or api key     | yes                                   | no                    | yes                      | yes                    | `member:manage`                 | organization integration present                        |
| `/api/v1/organizations/:id/members/:userId` | DELETE | tenantAuthorized                 | yes or api key     | yes                                   | no                    | yes                      | yes                    | `member:manage`                 | organization integration present                        |
| `/api/v1/organizations/:id/presence`        | GET    | tenantAuthorized                 | yes or api key     | yes                                   | no                    | yes                      | yes                    | `presence:view`                 | presence hook/E2E partial                               |
| `/api/v1/projects`                          | POST   | tenantAuthorized                 | yes or api key     | yes                                   | no                    | yes                      | yes                    | `project:manage`                | project integration/e2e present                         |
| `/api/v1/projects`                          | GET    | tenantAuthorized                 | yes or api key     | yes                                   | no                    | yes                      | yes                    | `project:view`/`project:manage` | project integration/e2e present                         |
| `/api/v1/projects/:id`                      | PUT    | tenantAuthorized                 | yes or api key     | yes                                   | no                    | yes                      | yes                    | `project:manage`                | project integration/e2e present                         |
| `/api/v1/webhooks`                          | CRUD   | tenantAuthorized                 | yes or api key     | yes                                   | no                    | yes                      | yes                    | `webhook:manage`                | webhook integration/e2e present                         |
| `/api/v1/organizations/:id/restore`         | POST   | authorized                       | yes                | api key denied by nested user-session | yes for nested route  | optional                 | yes                    | `admin:manage`                  | organization restore tests present                      |
| `/api/v1/organizations/:id/hard`            | DELETE | authorized                       | yes                | api key denied by nested user-session | yes for nested route  | optional                 | yes                    | `admin:manage`                  | organization hard-delete tests present                  |
| `/api/v1/permissions/*`                     | mixed  | authorized                       | yes or api key     | yes                                   | no                    | optional                 | yes                    | `admin:manage`                  | permission integration/e2e present                      |
| `/api/v1/access-rights*`                    | mixed  | authorized + nested optional org | yes or api key     | yes                                   | no                    | optional                 | yes                    | `admin:manage`                  | access integration/e2e present                          |
| `/api/v1/roles*`                            | mixed  | authorized                       | yes or api key     | yes                                   | no                    | optional                 | yes                    | `admin:manage`                  | role integration/e2e present                            |
| `/api/v1/users` admin paths                 | mixed  | authorized                       | yes or api key     | yes                                   | no                    | optional                 | yes                    | `admin:manage`                  | user e2e/integration present                            |
| `/api/v1/audit-logs/search`                 | POST   | authorized                       | yes or api key     | yes                                   | no                    | optional                 | yes                    | `admin:manage`                  | audit integration/e2e present                           |
| `/api/v1/upload/files/*any`                 | ANY    | upload                           | yes                | no                                    | yes                   | no                       | no                     | none                            | tus integration/e2e present                             |

## Coverage Gaps

### Stronger Signals Present

- auth lifecycle
- tenant isolation
- API key lifecycle
- permission/role/access orchestration
- project routes
- audit export
- TUS integration/e2e
- realtime SSE/WS e2e

### Weak or Needs Confirmation

- exact SSO route lifecycle coverage
- exact `organizations/slug/:slug` route coverage
- explicit route test for `auth/ticket`
- explicit route test for nested `RequireUserSession()` behavior on org delete/admin restore/hard-delete with API key
- explicit route test proving `authenticated` group blocks API-key-only access across representative endpoints
- explicit route test proving `authorized` group falls back to `global` domain only when intended
- explicit route test for upload route rejecting API-key-only access

## Main Review Rules

- never add user-session-only route to `tenantAuthorized` or `authorized` without checking API-key implications
- never add org-scoped route outside `tenantAuthorized` unless behavior is intentionally non-tenant
- never assume API key means allowed; check `RequireUserSession()` and explicit scopes
- never assume JWT means enough; authenticated truth includes Redis-backed session verification
- never change route paths without checking Casbin policy path semantics

## Suggested Next Tests

1. representative route tests for each group proving middleware semantics
2. API-key-denied tests for `authenticated` group
3. org-domain Casbin tests for `tenantAuthorized`
4. `OptionalOrganization()` admin path tests proving intended `global` vs org behavior
5. upload route tests for session-only access and avatar mutation ownership
