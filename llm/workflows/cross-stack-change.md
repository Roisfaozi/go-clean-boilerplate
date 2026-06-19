# Cross-Stack Change Workflow

## Purpose

Workflow ini untuk perubahan yang menyentuh backend Go plus frontend consumers, proxy layers, shared types, atau shared packages.

Fokus utamanya adalah producer-consumer sync.

## Use when

- backend API contract changes and either frontend app can consume it
- shared types or proxy behavior changes
- frontend and backend need coordinated rollout
- route, cookie, auth, or tenant semantics affect app behavior

## Read first

1. `AGENTS.md`
2. `llm/cache/api-contracts.md`
3. `llm/cache/frontend-map.md`
4. `llm/cache/frontend-proxy-system.md`
5. `llm/workflows/api-endpoint.md`
6. `llm/conventions/typescript.md`

## Live code to inspect

- backend route/controller/usecase files
- `apps/web/src/app/api/v1/[...path]/route.ts`
- `apps/client/app/routes/api-proxy.ts`
- `packages/api-types/*`
- frontend consumer files under owning app

## Workflow phases

### Phase 1 — Identify producer and consumers

State:

- backend producer endpoint or payload owner
- `apps/web` consumer or proxy path
- `apps/client` consumer or proxy path
- shared type owner if any

### Phase 2 — Define contract delta

Write exact change in:

- request params/body
- response shape
- error shape
- auth/tenant expectations
- proxy forwarding assumptions when relevant

### Phase 3 — Patch in order

Preferred order:

1. backend contract owner
2. shared types
3. proxies
4. frontend consumers

### Phase 4 — Verify both sides

- backend route or package tests
- app typecheck or focused consumer checks
- browser/E2E only when actual flow confidence is needed

## Review checklist

- both active apps checked when relevant
- proxy behavior still matches auth/cookie/header needs
- shared types are not stale
- backend docs or swagger updated when public contract changed

## Stop conditions / needs confirmation

- one consumer path cannot be identified
- producer and consumer verification pair is missing
- backend and frontend disagree on source of truth for contract
