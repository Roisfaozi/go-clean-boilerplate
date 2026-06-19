# Authentication System

## Purpose

Durable map for backend authentication and session behavior in this repo.

## Runtime truth

- `internal/router/router.go` wires authenticated HTTP route groups.
- `internal/middleware/auth_middleware.go` owns bearer/cookie token extraction and Redis-backed session verification.
- `internal/modules/auth` owns login/register/token/password/SSO/WebSocket ticket flows.
- `internal/config/app.go` wires JWT manager, auth module, ticket manager, SSO providers, Redis, audit, worker, WS/SSE managers, and organization repository into auth.

## Request authentication flow

1. `apiKeyMiddleware.Authenticate()` runs first on protected groups and may set API-key identity.
2. `authMiddleware.ValidateToken()` skips JWT token validation when API-key auth is already present.
3. JWT token is read from `Authorization: Bearer ...` or `access_token` cookie.
4. `AuthUseCase.ValidateAccessToken()` parses claims.
5. `AuthUseCase.Verify(ctx, userID, sessionID)` verifies session state; invalid/revoked sessions are rejected.
6. user/session/role/username are placed into Gin context and request context via `pkg/authcontext`.

## WebSocket authentication

- `/api/v1/ws` uses `authMiddleware.ValidateWebSocketToken()`.
- WebSocket auth requires `ticket` query parameter, not raw access token.
- Ticket validation goes through `pkg/ws.TicketManager` and sets user/session/role/username context.
- Ticket context can include `organization_id`.

## Public/authenticated auth routes

- Public auth routes are registered under `/api/v1/auth` in `internal/router/router.go` public group.
- Authenticated auth routes are registered under `/api/v1/auth` in authenticated group.
- Public group gets public/critical rate limiters depending on route.

## Hard rules

- Do not treat JWT parsing alone as authenticated session success.
- Do not bypass Redis-backed session validation for protected JWT routes.
- Do not accept raw access tokens for WebSocket route unless live code intentionally changes that boundary.
- Keep auth/session checks in middleware/usecase boundaries, not frontend-only checks.

## Tests and evidence paths

- `internal/modules/auth/test/*`
- `internal/modules/auth/repository/token_repository_test.go`
- `internal/middleware/auth_middleware.go`
- `pkg/ws/ticket_manager_test.go`
