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
4. update UI screens/components in each active app that consumes the change.
5. verify cookie/header/session behavior if auth is affected.
6. run backend and frontend checks.
7. update `llm/cache/api-contracts.md` if durable contract ownership changed.

## Verification

- backend route/middleware/module tests.
- `apps/web` typecheck/build/lint if `apps/web` changed.
- `apps/client` typecheck/build/E2E if `apps/client` changed.
- do not treat `apps/client` lint as meaningful until lint script is real.
- if only one frontend app is affected, state why the other is unaffected.
