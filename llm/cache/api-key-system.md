# API Key System

## Purpose

Durable map for API-key authentication, scope enforcement, organization-scoped API keys, and protected endpoint scope rules.

## Runtime truth

- `internal/modules/api_key/module.go` wires repository, organization repository, user repository, Redis, controller, and usecase.
- `internal/modules/api_key/delivery/http/api_key_routes.go` registers `/api-keys` under authenticated + tenant-required routes.
- `internal/modules/api_key/usecase/api_key_usecase.go` owns create/list/revoke/authenticate behavior.
- `internal/middleware/api_key_middleware.go` handles request authentication and scope enforcement.

## Behavior surfaces

- create/list/revoke API keys for an organization
- authenticate API key from `X-API-Key`
- inject user/org/username/scopes into request context
- scope derivation from HTTP method/path
- explicit scope checks for sensitive endpoints

## Hard rules

- API-key identity is not enough; route scope must allow action.
- API-key routes remain organization-scoped.
- Do not bypass auth/session/tenant layering when API key is present.

## Tests and evidence paths

- `internal/modules/api_key/test/*`
- `internal/modules/api_key/usecase/api_key_usecase.go`
- `internal/modules/api_key/delivery/http/api_key_routes.go`
