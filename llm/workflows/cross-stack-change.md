# Cross-Stack Change Workflow

## Use when

- changing both backend Go behavior and active frontend apps
- changing API path, auth cookies/tokens, proxy behavior, or typed contract usage
- changing shared packages used by frontend apps and backend contracts

## Read first

- `llm/cache/api-contracts.md`
- `llm/cache/backend-map.md`
- `llm/cache/frontend-map.md`
- `llm/cache/module-map.md`
- `llm/conventions/typescript.md`
- relevant Go convention files

## Live code to inspect

- backend route/controller/usecase files
- frontend proxy files:
  - `apps/web/src/app/api/v1/[...path]/route.ts`
  - `apps/client/app/routes/api-proxy.ts`
- frontend API client/helpers:
  - `apps/web/src/lib/api`
  - `apps/client/app/lib/api`
- shared type package: `packages/api-types`
- affected feature screens/components

## Steps

1. establish backend contract from route/model/usecase first.
2. update generated/shared types if type contract changes.
3. update frontend API client/proxy usage.
4. update UI screens/components in each active app that consumes change.
5. verify cookie/header/session behavior if auth is affected.
6. run backend and frontend checks.
7. update `llm/cache/api-contracts.md` only in a separate committed documentation task, not while behavior is still uncommitted.

## Verification commands

- backend package tests or `pnpm go:test`
- affected frontend app checks: `pnpm --filter casbin-web typecheck`, `pnpm --filter casbin-web build`, `pnpm --filter casbin-client typecheck`
- integration/E2E when auth, proxy, cookies, or tenant routing changed: `pnpm go:test-integration`, `pnpm go:test-e2e`

## Review checklist

- backend remains source of truth for contract semantics
- both active apps audited for affected consumer paths
- proxy/header/cookie forwarding preserved
- shared types updated before downstream UI assumptions
- no frontend-only auth bypass introduced

## Stop conditions / needs confirmation

- backend and frontend expectations disagree on payload semantics
- one active app has hidden consumer path that cannot be verified locally
- requested change conflicts with current proxy architecture
