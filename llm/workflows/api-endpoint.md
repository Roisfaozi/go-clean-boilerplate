# API Endpoint Workflow

## Purpose

Workflow ini untuk menambah atau mengubah backend HTTP endpoint dengan route stratum, scope, contract, and consumer sync yang benar.

## Use when

- adding or changing endpoint Gin
- moving endpoint between route groups
- changing request/response contract consumed by frontend or external clients
- changing API-key scope, tenant, or Casbin route semantics

## Read first

1. `AGENTS.md`
2. `llm/cache/backend-map.md`
3. `llm/cache/api-contracts.md`
4. `llm/cache/domain-rules.md`
5. `llm/conventions/golang.md`
6. `llm/conventions/testing.md`

## Live code to inspect

- `internal/router/router.go`
- relevant module route file under `internal/modules/*/delivery/http/*routes.go`
- target controller/usecase/repository/model files
- `internal/middleware/*` if auth, tenant, Casbin, or API-key behavior changes
- `docs/swagger.yaml`, `docs/swagger.json`, `docs/docs.go`
- frontend proxy/client files if consumed by frontend

## Route-group decision

Choose one intentionally:

- `public`
- `authenticated`
- `tenantAuthorized`
- `authorized`
- `upload`

Do not pick route group by convenience.

## Workflow phases

### Phase 1 — Define contract and owner

State:

- method and path
- request shape
- response and error shape
- owning module
- intended consumer(s)

### Phase 2 — Choose route stratum and scope

Confirm:

- auth/session rules
- API-key scope behavior
- tenant or organization context
- Casbin requirement or lack of it

### Phase 3 — Patch by layer

- route registration
- request/response models and validation tags
- controller parsing and response
- usecase logic
- repository or schema if needed

### Phase 4 — Sync documentation and consumers

- update Swagger generation path when public contract changes
- update `packages/api-types`, proxies, and app consumers when relevant

### Phase 5 — Verify narrow then broad

- narrow backend package test first
- route/auth changes: broader integration when needed
- public contract/doc changes: `pnpm go:docs`
- frontend consumers changed: app typecheck/build for touched app

## Review checklist

- route registered exactly once
- route protection matches intended stratum
- API-key scope and tenant/Casbin layering preserved
- request/response model matches actual handler output
- swagger artifacts updated when contract is public and documented

## Stop conditions / needs confirmation

- route belongs to more than one plausible security stratum and intent is unclear
- existing frontend consumers conflict with requested contract change
- endpoint implies migration or background side effect not described in request
