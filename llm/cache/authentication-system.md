# Authentication System

## Purpose

Durable map for backend authentication, session verification, SSO, WebSocket ticket auth, and protected-route identity behavior in this repo.

## Runtime truth

- `internal/modules/auth` owns login, register, token, password reset, verification, SSO, logout, and ticket flows.
- `internal/middleware/auth_middleware.go` owns bearer and cookie token extraction plus Redis-backed session verification.
- `internal/config/app.go` wires JWT manager, auth module, ticket manager, SSO providers, Redis, audit, worker, WS/SSE managers, and organization repository into auth.
- `internal/router/router.go` defines public and authenticated auth route registration.

## Protected request authentication flow

On protected route groups the effective flow is:

1. `apiKeyMiddleware.Authenticate()` may establish API-key actor first.
2. `authMiddleware.ValidateToken()` runs next and skips JWT token validation when API-key auth is already present.
3. JWT token is read from `Authorization: Bearer ...` or `access_token` cookie.
4. `AuthUseCase.ValidateAccessToken()` parses claims.
5. `AuthUseCase.Verify(ctx, userID, sessionID)` checks backing session state in Redis or session store semantics.
6. user, session, role, and username are placed into Gin context and request context via `pkg/authcontext`.

This means parsed JWT is not enough by itself for protected JWT routes.

## Public auth surfaces

Public auth routes under `/api/v1/auth` include:

- login
- refresh
- forgot password
- reset password
- verify email
- register
- SSO login
- SSO callback

`internal/router/router.go` gives login special critical rate limiting when configured.

## Authenticated auth surfaces

Authenticated auth routes under `/api/v1/auth` include:

- logout
- ticket creation for WebSocket
- resend verification
- current-user `me`

These run under authenticated route layering, not public route layering.

## WebSocket and SSE auth behavior

- `/api/v1/events` uses `authMiddleware.ValidateToken()`.
- `/api/v1/ws` uses `authMiddleware.ValidateWebSocketToken()`.
- WebSocket auth requires `ticket` query parameter, not raw access token by default.
- Ticket validation goes through `pkg/ws.TicketManager` and can carry user, session, role, username, and optionally `organization_id` context.

## SSO and token lifecycle notes

- SSO providers are wired in app config and auth module construction.
- Auth flow can interact with frontend callbacks and proxy behavior, especially in `apps/web` auth routes.
- Session revocation semantics matter for logout and any flow that depends on current active session state.

## Coupling to other systems

- API-key middleware can coexist with protected routes and changes effective auth path.
- user status middleware follows auth on protected groups.
- tenant and Casbin checks assume auth identity is already established on protected tenant/admin routes.
- WebSocket ticket flow depends on auth usecase plus realtime stack.

## Hard rules

- Do not treat JWT parsing alone as authenticated session success.
- Do not bypass Redis-backed session validation for protected JWT routes.
- Do not accept raw access tokens for `/ws` unless live code intentionally changes that trust boundary.
- Keep auth and session checks in middleware or usecase boundaries, not frontend-only checks.
- Auth changes can affect browser cookie flows, proxy forwarding, and ticket behavior; verify those consumers when relevant.

## Verification and evidence paths

- `internal/modules/auth/test/*`
- `internal/modules/auth/repository/token_repository_test.go`
- `internal/middleware/auth_middleware.go`
- `pkg/ws/ticket_manager_test.go`
- `internal/router/router.go`
- `internal/config/app.go`
