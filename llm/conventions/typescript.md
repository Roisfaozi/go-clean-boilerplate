# TypeScript Conventions

## Active apps

- `apps/web`: Next.js App Router with strict TypeScript.
- `apps/client`: React Router + Vite with strict TypeScript.
- Both are active production-bound surfaces and should not be treated as legacy.

## Import/path conventions

- `apps/web` uses `~/*` mapped to `src/*`.
- `apps/client` uses `~/*` and `@/*` mapped to `app/*`.
- Prefer workspace packages already used by app package.json: `@casbin/api-types`, `@casbin/hooks`, `@casbin/ui`, `@casbin/utils`.

## Project shape

`apps/web`:

- `src/app`: Next.js App Router pages, layouts, route handlers, server actions.
- `src/components`: auth, dashboard, UI, role/user/access components.
- `src/lib/api`: backend client helpers.
- `src/lib/server`: server-only auth helpers.
- `src/hooks`, `src/stores`, `src/types`, `src/locales`: shared frontend support.

`apps/client`:

- `app/routes.ts`: route registry and navigation contract.
- `app/pages`: page-level screens.
- `app/features`: domain feature areas.
- `app/components`: reusable UI by concern.
- `app/lib`: API, upload, realtime, email helpers.
- `app/hooks`, `app/stores`: reusable state/behavior.

## Backend boundary

- Keep backend API base/path handling centralized in proxy or API client helpers.
- `apps/web` proxy: `apps/web/src/app/api/v1/[...path]/route.ts`.
- `apps/client` proxy: `apps/client/app/routes/api-proxy.ts`.
- Do not duplicate auth cookie/header forwarding logic in random components.
- For API contract changes, update shared `packages/api-types` if used by the feature.

## UI and composition

- Prefer existing shared UI primitives from `packages/ui` and app-local UI folders.
- Keep feature-specific UI inside feature folders unless it is genuinely reusable.
- Use route-level/page-level components for composition and keep transport/API helpers in `lib`.

## Verification

- `apps/web`: `lint`, `typecheck`, `build` scripts exist in package.
- `apps/client`: `typecheck`, `build`, and Playwright E2E scripts exist in package.
- `apps/client` currently has placeholder lint script (`echo 'lint not configured yet'`), so do not treat client lint as meaningful verification until repo tooling changes.
- Cross-stack changes should verify backend plus affected frontend app, not frontend alone.
