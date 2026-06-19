# API Key System

## Purpose

Durable map for API-key authentication, scope enforcement, organization-scoped API keys, and protected endpoint scope rules.

## Runtime truth

- Module root: `internal/modules/api_key/`
- Wiring entry: `internal/modules/api_key/module.go`
- Routes: `internal/modules/api_key/delivery/http/api_key_routes.go`
- Usecase: `internal/modules/api_key/usecase/api_key_usecase.go`
- Middleware boundary: `internal/middleware/api_key_middleware.go`

## Route ownership

- API-key management routes are registered through `api_keyHttp.RegisterApiKeyRoutes(authenticated, ...)` under the `authenticated` group in `internal/router/router.go`.
- The authenticated group still includes token validation, auto scope behavior, user-session requirement, and user-status middleware.
- Some other protected routes add explicit scope strings on top of auto-scope behavior, for example project routes.

## Behavior surfaces

- create, list, and revoke organization API keys
- authenticate API key from `X-API-Key`
- inject user, organization, username, and scopes into request context
- derive or enforce route scopes
- combine API-key actor with protected route semantics

## Scope semantics

- API-key identity alone is not enough.
- Route scope must explicitly allow the action.
- Tenant-aware routes can still require organization semantics beyond raw key presence.
- Admin-style routes can require explicit scope such as `admin:manage`.

## Hard rules

- Do not bypass auth, session, tenant, or scope layering just because API key is present.
- API-key management remains organization-scoped.
- Treat route scope changes as security changes, not convenience-only refactors.

## Verification and evidence paths

- `internal/modules/api_key/test/*`
- `internal/modules/api_key/usecase/api_key_usecase.go`
- `internal/modules/api_key/delivery/http/api_key_routes.go`
- `internal/middleware/api_key_middleware.go`
- `internal/router/router.go`
