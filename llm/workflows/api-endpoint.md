# API Endpoint Workflow

## Use when

- adding or changing backend HTTP endpoints
- changing auth, tenant, API key, or Casbin behavior on routes
- changing endpoint contract consumed by `apps/web`, `apps/client`, or `packages/api-types`

## Read first

- `llm/cache/architecture.md`
- `llm/cache/backend-map.md`
- `llm/cache/api-contracts.md`
- `llm/cache/domain-rules.md`
- `llm/conventions/golang.md`
- `llm/conventions/testing.md`

## Live code to inspect

- `internal/router/router.go`
- relevant module route file under `internal/modules/*/delivery/http/*routes.go`
- target controller/usecase/repository/model files
- `internal/middleware/*` if auth/tenant/casbin/api-key behavior changes
- `docs/swagger.yaml` / `docs/swagger.json` / `docs/docs.go`
- frontend proxy/client files if consumed by frontend

## Route-group decision

Choose one intentionally:

- public: no auth; only safe public auth/invitation style flows.
- authenticated: API-key/JWT/session/status checks, but no required tenant/Casbin route policy.
- tenantAuthorized: auth + tenant org + Casbin policy.
- authorized: admin-style scope + optional tenant + Casbin policy.
- upload: TUS route with auth/status and upload-specific handler.

## Steps

1. define method/path/request/response and owning module.
2. select route group and required API-key scopes.
3. add or update request/response models and validation tags.
4. update controller and usecase behavior.
5. update repository or schema if needed.
6. update Swagger comments/docs generation if contract should appear in OpenAPI.
7. update frontend API types/client/proxy use if affected.
8. run targeted tests and broader route tests if authorization changed.

## Verification commands

- narrow backend package test for owning module
- route/auth changes: `pnpm go:test-integration`
- public contract/doc changes: `pnpm go:docs`
- frontend consumers changed: app typecheck/build for touched app

## Review checklist

- route registered exactly once
- route protection matches intended stratum
- API-key scope and tenant/Casbin layering preserved
- request/response model matches actual handler output
- Swagger artifacts updated when contract is public and documented

## Stop conditions / needs confirmation

- route belongs to more than one security stratum and product intent is unclear
- existing frontend consumers conflict with requested contract change
- endpoint implies migration or background side effect not described in request
