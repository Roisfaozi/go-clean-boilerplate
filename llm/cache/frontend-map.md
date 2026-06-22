# Frontend Map

## Active frontend apps

### `apps/web`

Stack:

- Next.js 16
- React 19
- TypeScript
- Tailwind CSS

Entrypoints and key files:

- `apps/web/next.config.mjs`
- `apps/web/src/app/[locale]/layout.tsx`
- `apps/web/src/app/[locale]/page.tsx`
- `apps/web/src/app/api/v1/[...path]/route.ts`
- `apps/web/src/app/api/auth/callback/route.ts`
- `apps/web/src/app/api/auth/logout/route.ts`

Notable route tree:

- public/auth flows under `src/app/[locale]/(auth)`
- dashboard flows under `src/app/[locale]/dashboard`
- server-side auth proxy callbacks under `src/app/api/auth/*`
- backend proxy pass-through under `src/app/api/v1/[...path]/route.ts`

Structure:

- `src/app`: App Router pages, layouts, route handlers, server actions.
- `src/components`: UI, dashboard, auth, role/user/access components.
- `src/lib/api`: backend API client helpers.
- `src/lib/server`: server-only auth utilities.
- `src/hooks`, `src/stores`, `src/types`, `src/locales`: supporting frontend layers.

Additional frontend quality notes:

- shared providers and widgets include websocket, audit-stream, notification, and dashboard-specific components.
- the app mixes server actions and proxy routes, so backend-boundary changes should check both route handlers and client helpers.

Backend boundary:

- `NEXT_PUBLIC_API_URL` controls backend base URL.
- `src/app/api/v1/[...path]/route.ts` proxies API paths to backend.
- auth actions and callback/logout routes coordinate backend tokens/cookies.
- `NEXT_PUBLIC_WS_URL` controls WebSocket URL used by shared providers.

Runtime behavior notes:

- `access_token` cookie is promoted to bearer auth for backend proxy calls.
- cookies are forwarded to backend so session-based auth survives proxying.
- safe response headers are preserved for JSON/text responses.

### `apps/client`

Stack:

- React Router 7
- Vite
- React 19
- TypeScript
- Playwright E2E

Entrypoints and key files:

- `apps/client/react-router.config.ts`
- `apps/client/vite.config.ts`
- `apps/client/app/root.tsx`
- `apps/client/app/routes.ts`
- `apps/client/app/routes/api-proxy.ts`

Notable route tree:

- auth/login/register/reset flows live in `app/pages/auth`
- landing and marketing screens are in `app/pages` and `app/components/landing`
- dashboard/admin-like app surfaces are in `app/features/*`
- route config is generated from flat route definitions in `app/routes.ts`

Structure:

- `app/routes.ts`: route registry.
- `app/pages`: page-level screens.
- `app/features`: domain feature pages/components for users, roles, organizations, projects, permissions, resources, endpoints, audit logs.
- `app/components`: shared components grouped by UI concern.
- `app/lib`: API, realtime, upload, email helpers.
- `app/stores`, `app/hooks`: state and reusable behavior.

Additional frontend quality notes:

- the app has a rich showcase layer (`pages/showcase`) that exercises design system components.
- features are organized by domain, not by transport layer, which makes the route registry the key navigation source.

Backend boundary:

- `app/routes/api-proxy.ts` proxies `api/v1/*` to backend using `NEXT_PUBLIC_API_URL` fallback.
- `app/lib/api/client.ts` centralizes API client behavior.
- TUS upload client lives under `app/lib/upload`.

Runtime behavior notes:

- request proxy keeps method/body/headers and streams backend body back to callers.
- Set-Cookie is preserved explicitly to keep backend sessions usable from the frontend app.

## Shared packages

Root workspace includes shared packages:

- `packages/api-types`
- `packages/hooks`
- `packages/ui`
- `packages/utils`

These are consumed by frontend apps through workspace dependencies.

Shared package role notes:

- `api-types` is for contract typing.
- `hooks` holds reusable UI/data hooks.
- `ui` provides shared design-system primitives.
- `utils` holds general frontend helpers and cross-app utilities.
