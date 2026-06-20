# Cross-Stack Change Workflow

## Purpose

Primary workflow for changes that touch Go backend plus frontend consumers, proxy layers, shared types, or shared packages.

Goal: keep producer and consumer in sync across this monorepo.

## Use When

- backend API contract changes and `apps/web` or `apps/client` can consume it
- proxy behavior changes
- shared types in `packages/api-types` change
- auth/cookie/tenant semantics affect frontend behavior
- one logical fix spans backend plus frontend surface

## Do Not Use When

- task is strictly backend with no active frontend consumer
- task is strictly frontend styling with no contract/runtime change
- task is docs-only

## Required Read Order

1. `AGENTS.md`
2. `llm/cache/api-contracts.md`
3. `llm/cache/frontend-map.md`
4. `llm/cache/frontend-proxy-system.md`
5. `llm/cache/domain-rules.md`
6. `llm/workflows/api-endpoint.md` if route contract is changing
7. `llm/conventions/typescript.md`
8. `llm/workflows/benchmarking.md` if proxy/API/runtime path performance can change

## Live Code to Inspect

- backend route/controller/usecase files
- `internal/router/router.go`
- `apps/web/src/app/api/v1/[...path]/route.ts`
- `apps/client/app/routes/api-proxy.ts`
- `packages/api-types/*`
- frontend consumer files under owning app

## Workflow Steps

### Step 1 — Identify Producer and All Consumers

State explicitly:

- backend producer endpoint or payload owner
- `apps/web` consumer or proxy owner
- `apps/client` consumer or proxy owner
- shared type owner if any

Do not assume one app is inactive just because current page is not open.

### Step 2 — Define Contract Delta

Write exact change in:

- request params/body/query
- response shape
- error envelope
- auth/tenant expectations
- proxy forwarding semantics if relevant

### Step 3 — Patch In Stable Order

Before patching existing runtime hot path, perform benchmark audit when change can affect latency, allocations, query cost, or proxy overhead.

Default order:

1. backend contract owner
2. shared types
3. proxies
4. frontend consumers
5. UI state handling

### Step 4 — Re-check Boundary States

For contract changes, verify:

- loading
- empty
- error
- unauthorized/forbidden
- tenant mismatch if relevant
- offline/backend unavailable handling through proxy behavior

## Common Mistakes

- patching backend and only one app
- updating consumer code but forgetting proxy or shared type
- assuming generated types are source of truth over route wiring
- forgetting cookie/header/auth implications

## Verification

### Minimum

- narrow backend package test or route verification
- owner app typecheck

### Add When Needed

- second app typecheck if shared consumer contract changed
- browser/manual flow
- E2E when auth, tenant, cookie, or route lifecycle changed
- `pnpm go:docs` if public Swagger contract changed

## Stop Conditions

Stop and mark `needs confirmation` if:

- one consumer path cannot be identified
- producer and consumer disagree on source of truth
- contract delta affects both apps but one owner path is unverified

## Completion Output

Report:

- producer changed
- consumers changed
- proxy/shared-type impact
- verification run per side
