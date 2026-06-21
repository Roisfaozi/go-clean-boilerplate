# TypeScript Conventions

## Purpose

Guide for frontend ownership, shared package usage, proxy boundaries, and TypeScript discipline across `apps/web`, `apps/client`, and workspace packages.

## Active apps

- `apps/web`: Next.js App Router with strict TypeScript
- `apps/client`: React Router plus Vite with strict TypeScript
- both are active surfaces and should not be treated as legacy or secondary by default

## Ownership first

Before editing TypeScript code, prove which surface owns behavior:

- `apps/web`
  - App Router pages, layouts, route handlers, server actions
- `apps/client`
  - route registry, page screens, domain feature folders
- `packages/*`
  - shared types, hooks, UI primitives, utils only when reuse is real

If owner is unclear, stop and resolve owner before refactor.

## Import and path conventions

- `apps/web` uses `~/*` mapped to `src/*`
- `apps/client` uses `~/*` and `@/*` mapped to `app/*`
- prefer workspace packages already used by app `package.json`:
  - `@casbin/api-types`
  - `@casbin/hooks`
  - `@casbin/ui`
  - `@casbin/utils`

## Project shape guidance

### `apps/web`

- `src/app`
  - App Router pages, layouts, route handlers, server actions
- `src/components`
  - auth, dashboard, UI, role/user/access components
- `src/lib/api`
  - backend client helpers
- `src/lib/server`
  - server-only auth helpers

### `apps/client`

- `app/routes.ts`
  - route registry and navigation contract
- `app/pages`
  - page-level screens
- `app/features`
  - domain feature areas
- `app/components`
  - reusable UI by concern
- `app/lib`
  - API, upload, realtime, email helpers

## Backend boundary rules

- keep backend API base/path handling centralized in proxy or API client helpers
- `apps/web` proxy is `apps/web/src/app/api/v1/[...path]/route.ts`
- `apps/client` proxy is `apps/client/app/routes/api-proxy.ts`
- do not duplicate auth cookie or header forwarding logic inside random components
- for API contract changes, update `packages/api-types` when feature actually uses shared contract typing

## UI and composition rules

- prefer existing shared UI primitives from `packages/ui` and app-local UI folders
- keep feature-specific UI inside feature folders unless genuinely reusable
- use route-level or page-level components for composition
- keep transport/API helpers in `lib`, not embedded in arbitrary UI components

## State and UX expectations

Relevant UI changes should check for:

- loading
- empty
- error
- success
- auth-expired
- tenant-switch or org-sensitive state where relevant

Frontend visibility is not authorization. Route hiding or disabled UI is not substitute for backend protection.

## Cross-stack discipline

- backend contract change means checking producer and consumer, not just one side
- if both apps can consume changed path, audit both
- update shared types, proxies, and consuming hooks/components deliberately

## Verification expectations

- `apps/web`: `lint`, `typecheck`, `build` scripts exist and are meaningful
- `apps/client`: `typecheck`, `build`, and Playwright E2E scripts exist
- `apps/client` lint runs Biome; run `typecheck` separately for TypeScript validation
- cross-stack changes should verify backend plus affected frontend app, not frontend alone

## Common mistakes

- editing wrong app because feature names overlap
- promoting code to shared package too early
- duplicating proxy or client behavior in component layer
- changing backend contract without updating proxies or shared types
