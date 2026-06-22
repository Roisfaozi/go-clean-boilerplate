# API Endpoint Workflow

## Purpose

Primary workflow for adding or changing HTTP endpoints in this repo.

This workflow exists to preserve route strata, auth layering, API-key scope behavior, tenant context, Casbin enforcement, Swagger-visible contracts, and frontend consumer sync.

## Use When

- adding new backend HTTP route
- changing request/response shape of existing endpoint
- moving endpoint between route groups
- changing route protection or API-key scope behavior
- changing frontend-consumed API surface

## Do Not Use When

- task is purely internal Go logic with no endpoint contract change
- task is schema-only change with no route impact

## Required Read Order

1. `AGENTS.md`
2. `llm/cache/api-contracts.md`
3. relevant domain cache file
4. `internal/router/router.go`
5. target module route/controller/usecase files
6. `llm/workflows/cross-stack-change.md` if frontend consumers exist

## Live Code to Inspect

- `internal/router/router.go`
- target `*_routes.go`
- target controller/usecase/model files
- middleware files for auth, tenant, API key, Casbin
- frontend proxies and consumer files if active contract consumer exists

## Workflow Steps

### Step 1 — Choose Correct Route Stratum

Decide whether endpoint belongs in:

- `public`
- `authenticated`
- `tenantAuthorized`
- `authorized`
- upload or other special path

Do not pick by convenience. Pick by runtime security semantics.

### Step 2 — Confirm Protection Stack

State explicit expectations for:

- JWT/session validation
- API-key auth behavior
- required or forbidden API-key user-session semantics
- tenant/org context
- Casbin requirement or lack of it
- explicit scope override if needed

### Step 3 — Patch By Layer

Preferred order:

- route registration
- request/response models and validation tags
- controller parsing and response
- usecase logic
- repository/schema if needed

### Step 4 — Sync Consumers and Docs

If endpoint is frontend-consumed or public contract:

- update proxies or app consumers
- update shared types if used
- update Swagger path via `pnpm go:docs` when appropriate

## Common Mistakes

- putting route in wrong security group
- forgetting API-key scope override on sensitive resource
- changing controller output but not response model/docs
- assuming frontend does not consume route

## Verification

### Minimum

- narrow backend package or route test

### Add When Needed

- auth/route integration check when middleware layering changed
- `pnpm go:docs` when Swagger-visible contract changed
- owner app typecheck/build if frontend consumer changed

## Stop Conditions

Stop and mark `needs confirmation` if:

- route belongs to more than one plausible security stratum and intent unclear
- existing consumers conflict with requested contract
- endpoint implies migration or async side effect not mentioned in request

## Completion Output

Report:

- route stratum chosen and why
- auth/tenant/API-key/Casbin behavior
- consumer/docs impact
- verification run
