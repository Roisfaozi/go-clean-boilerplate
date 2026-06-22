# API Contracts

## Contract sources

Confirmed contract sources:

- Swagger/OpenAPI generated files under `docs/swagger.json`, `docs/swagger.yaml`, `docs/docs.go`.
- Backend route registration in `internal/router/router.go` and module route files.
- Frontend proxy and API client paths in `apps/web/src/app/api/v1/[...path]/route.ts`, `apps/client/app/routes/api-proxy.ts`, and frontend `lib/api` folders.
- Shared package `packages/api-types` exists for frontend-facing types.

Additional contract-like sources that need caution:

- `documentation/api/AI_STREAMING_CONTRACT.md` exists as a design/contract document, but it is not currently confirmed as a wired backend route in live code.

Do not invent API contracts from UI labels. Prefer backend routes and generated Swagger.

## Backend base

Primary backend prefix:

- `/api/v1`

Common top-level endpoints:

- `GET /api/v1/health`
- `GET /api/v1/docs/*any`
- `GET /api/v1/events`
- `GET /api/v1/ws`

Important non-CRUD upload surface:

- `ANY /api/v1/upload/files/*`

## Auth boundary

Public auth endpoints live under `/api/v1/auth`:

- login
- refresh
- forgot/reset password
- verify email
- register
- SSO login/callback

Authenticated auth endpoints include:

- logout
- WebSocket ticket
- resend verification
- current user (`me`)

## Module endpoint families

Observed route families in backend route files:

- `/users`
- `/organizations`
- `/stats`
- `/permissions`
- `/access-rights`
- `/endpoints`
- `/roles`
- `/projects`
- `/api-keys`
- `/audit-logs`
- `/webhooks`

## Frontend proxy boundary

`apps/web`:

- `apps/web/src/app/api/v1/[...path]/route.ts` proxies `/api/v1/*` style traffic to backend URL.
- It forwards bearer auth from `access_token` cookie to `Authorization` header when present.
- It forwards cookies and selected safe response headers.
- It returns a `BACKEND_OFFLINE` JSON response on fetch failure.

`apps/client`:

- route `api/v1/*` maps to `apps/client/app/routes/api-proxy.ts`.
- It strips `/api/v1` from the configured base URL and rebuilds the target path from incoming request pathname.
- It forwards cookies/headers, preserves `Set-Cookie`, and streams response body back to caller.
- It returns a `BACKEND_OFFLINE` JSON response on fetch failure.

## Authorization semantics

Backend routes may require combinations of:

- API key authentication
- JWT/session validation
- API key scopes
- active user status
- organization/tenant context
- Casbin policy allow

Route protection is defined in `internal/router/router.go` and module route files, not in generated frontend types.

## Contract confidence notes

- Swagger and route-registration files are high-confidence live contract sources.
- Frontend proxy files are high-confidence transport contract sources.
- `documentation/api/AI_STREAMING_CONTRACT.md` is currently best treated as planned or supporting documentation until a matching live route is found in backend routing.
