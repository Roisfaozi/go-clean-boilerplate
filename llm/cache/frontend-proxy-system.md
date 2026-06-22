# Frontend Proxy and Forms System

## Purpose

Durable map for frontend request boundaries in this monorepo:

- proxy surfaces in `apps/web` and `apps/client`
- cookie and auth forwarding behavior
- backend-offline behavior
- shared frontend contract risk when backend API changes
- common form submission patterns and their coupling to backend validation

Use this file before changing frontend API calls, auth cookies, proxy headers, form payload shape, or shared API contract assumptions.

## Primary source of truth

1. `llm/cache/frontend-map.md`
2. `llm/cache/api-contracts.md`
3. `apps/web/src/app/api/v1/[...path]/route.ts`
4. `apps/client/app/routes/api-proxy.ts`
5. target app API helper files
6. target form implementation files
7. `internal/router/router.go` if backend route/auth behavior changed

## Surface ownership

### `apps/web`

- framework: Next.js App Router
- proxy owner: `apps/web/src/app/api/v1/[...path]/route.ts`
- common auth/form owners also include `apps/web/src/lib/api/*`, server actions, and app-specific UI components

### `apps/client`

- framework: React Router 7
- proxy owner: `apps/client/app/routes/api-proxy.ts`
- common feature owners live under `app/features/*`, `app/routes/*`, and app-local api helpers

### Shared workspace

- `packages/api-types` for shared response/request typing
- `packages/hooks`, `packages/ui`, `packages/utils` for shared frontend behavior that may hide contract assumptions

## Proxy behavior by app

### `apps/web` proxy semantics

Current durable behavior from `apps/web/src/app/api/v1/[...path]/route.ts`:

- builds backend target from `NEXT_PUBLIC_API_URL` with local fallback
- reads `access_token` cookie
- sets `Authorization: Bearer ...` when cookie exists
- forwards cookies downstream
- returns structured `BACKEND_OFFLINE` JSON when backend fetch fails
- forwards only selected safe response headers back to browser

Implication:

- web app auth can break if cookie name or bearer-header assumption changes
- backend header and cookie behavior are part of contract, not transport detail only

### `apps/client` proxy semantics

Current durable behavior from `apps/client/app/routes/api-proxy.ts`:

- builds backend base URL from `NEXT_PUBLIC_API_URL` with local fallback
- forwards request headers and cookies
- forwards safe response headers including `Set-Cookie`
- streams backend response body back to caller
- returns `BACKEND_OFFLINE` JSON on fetch failure

Implication:

- client app may preserve more raw backend response shape than web app in some flows
- cookie propagation behavior can differ by app surface and needs audit before auth changes

## Contract risk model

Backend contract change is not done when controller compiles.

For this repo, change may require updating all of:

- backend controller/route
- frontend proxy path expectations
- app-local API helper
- form serializer or field mapping
- shared type package if used
- loading/error state UI

If route is consumed by both apps, both apps need audit even if only one currently exposes visible page.

## Form patterns proven in repo

### `apps/web` auth-style pattern

Observed durable pattern:

- `apps/web/src/components/auth/login-form.tsx` uses React Hook Form
- `zodResolver` performs schema validation
- submit path delegates to `apps/web/src/app/actions/auth.ts`
- flow handles loading, field errors, toast errors, auth store updates, permission fetch, and redirect

Meaning:

- backend auth response changes can break more than one UI component
- validation shape and redirect/permission initialization are part of runtime path

### `apps/client` CRUD-dialog pattern

Observed durable pattern:

- `apps/client/app/features/shared/crud-form-dialog.tsx` uses local values/errors state
- `zod` `safeParse` validates before submit
- shared dialog supports multiple field kinds
- caller owns actual submit and post-submit behavior

Meaning:

- server validation and form field schema can drift separately
- feature-specific caller code may still need changes even if shared dialog stays same

## Cross-stack coupling

Frontend proxy/form changes can silently intersect:

- auth cookie/session behavior
- API-key usage if frontend surfaces expose API-key management
- tenant-aware routes that depend on org context or route params
- upload/formdata handling
- backend validation and error envelope shape

Read with:

- `llm/cache/authentication-system.md`
- `llm/cache/api-contracts.md`
- `llm/cache/tenant-organization-system.md`
- `llm/cache/frontend-map.md`

## Known sharp edges

- `apps/web` and `apps/client` are both active; do not patch one and assume repo done.
- `apps/client` lint runs Biome; use typecheck and build/E2E too when runtime or browser behavior changes.
- proxy `BACKEND_OFFLINE` behavior is user-facing contract and should stay consistent unless intentionally redesigned.
- cookie/header forwarding lists are security-sensitive and should not be widened casually.

## Change checklist

Before editing frontend proxy or forms, prove these answers:

1. Which app owns route or UI: `apps/web`, `apps/client`, or both?
2. Does backend contract change response body, error envelope, auth cookie, or header semantics?
3. Is request JSON, form-data, server action, or proxied stream?
4. Which form schema or caller mapping must change with it?
5. Do shared packages encode old field names or types?
6. Is verification using strong command surface for owning app?

## Verification paths

Narrow first:

- owning app typecheck
- owning app build if route/component wiring changed
- target component or route tests if present

Broader when auth/proxy semantics changed:

- manual browser flow
- relevant E2E flow

## Hard rules

- Do not treat frontend route hiding as auth.
- Do not change backend contract without auditing proxy and form callers.
- Do not duplicate proxy helpers across apps without clear owner reason.
- Do not rely on `apps/client` lint as strong verification.
- Do not widen forwarded headers/cookies casually.
