---
name: cross-stack-change
description: Use when a change spans Go backend plus `apps/web`, `apps/client`, frontend proxies, shared API types, or shared packages.
---

# Cross Stack Change: Backend Contract to Frontend Surfaces

**Announce at start:** "I'm using the cross-stack-change skill to keep backend contracts and both frontend apps aligned."

## Read Order

1. `AGENTS.md`
2. `llm/cache/api-contracts.md`
3. `llm/cache/frontend-map.md`
4. `llm/cache/backend-map.md`
5. `llm/workflows/cross-stack-change.md`
6. backend route/controller/usecase
7. `apps/web/src/app/api/v1/[...path]/route.ts`
8. `apps/client/app/routes/api-proxy.ts`
9. `packages/api-types` if shape changes

## Workflow

### Phase 1 — Backend Truth

- Establish contract from live route/controller/usecase.
- Identify auth, cookie, token, tenant header, and error response behavior.

### Phase 2 — Consumer Inventory

Check both active apps:

- `apps/web` Next.js App Router/server actions/proxy
- `apps/client` React Router/proxy/client helpers
- shared `packages/*`

### Phase 3 — Patch Order

1. backend contract
2. shared types/helpers
3. frontend proxy/client
4. feature UI state/error handling

### Phase 4 — Verify

- backend targeted tests
- frontend typecheck/build for touched app
- integration/E2E for auth/cookie/proxy/request lifecycle changes

## Red Flags

- updating only one active frontend app when both consume route
- frontend-only auth/tenant enforcement
- changing payload shape without shared types/proxy update
- treating `apps/client` lint as strong verification

## Stop Conditions

- Stop and ask before destructive DB/schema/data operations not explicitly requested.
- Stop if live code contradicts `llm/cache/*`; live code wins, then document drift in `llm/tasks/`.
- Stop if route ownership, tenant boundary, or auth stratum is unclear.

## Completion Output

Report:

- files changed
- commands run and exact result
- verification skipped and exact blocker
- risks or follow-up work
