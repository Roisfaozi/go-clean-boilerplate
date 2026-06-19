# API Key System

## Purpose

Durable map for API-key authentication, scope enforcement, organization-scoped API keys, and protected endpoint scope rules.

## Runtime truth

- `internal/middleware/api_key_middleware.go` authenticates `X-API-Key`, injects identity, sets organization context, and enforces scopes.
- `internal/modules/api_key` owns API-key entity/model/repository/usecase/controller/routes/tests.
- `internal/router/router.go` applies API-key middleware before JWT validation on protected groups.

## Middleware behavior

- `Authenticate()` reads `X-API-Key`; if absent, request continues for JWT path.
- Authenticated API-key identity sets `user_id`, `organization_id`, `username`, API-key ID, and scopes.
- API-key auth sets DB organization context when organization ID exists.
- `RequireScopeAuto()` derives scope from route resource and method.
- `RequireScopes()` enforces any of specified scopes.
- `RequireAllScopes()` enforces all specified scopes.
- `RequireUserSession()` rejects API-key auth for endpoints requiring a user session.

## Scope mapping

- GET/HEAD -> `view`
- POST -> `create`
- PUT/PATCH -> `update`
- DELETE -> `delete`
- `*` grants all scopes.

## Hard rules

- API-key identity is not enough; scope must allow action where route requires it.
- Protected endpoint additions need an explicit API-key scope decision.
- Organization-scoped behavior must remain aligned with tenant context.

## Tests and evidence paths

- `internal/modules/api_key/test/*`
- `internal/middleware/api_key_middleware.go`
- `internal/router/router.go`
