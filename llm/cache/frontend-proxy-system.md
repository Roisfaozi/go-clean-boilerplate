# Frontend Proxy and Forms System

## Purpose

Durable map for active frontend API proxy surfaces, shared client boundaries, and form patterns.

## Active frontend surfaces

- `apps/web`: Next.js App Router, server actions, UI components, backend API proxy.
- `apps/client`: React Router 7, feature folders, route registry, backend API proxy.
- `packages/*`: shared types, hooks, UI, and utils.

## API proxy surfaces

### `apps/web`

- Proxy file: `apps/web/src/app/api/v1/[...path]/route.ts`.
- Builds backend URL from `NEXT_PUBLIC_API_URL` fallbacking to `http://127.0.0.1:8080/api/v1`.
- Reads `access_token` cookie and sets `Authorization: Bearer ...` when present.
- Explicitly forwards cookies and selected response headers.
- Returns `BACKEND_OFFLINE` JSON on backend fetch failure.

### `apps/client`

- Proxy file: `apps/client/app/routes/api-proxy.ts`.
- Builds backend base URL from `NEXT_PUBLIC_API_URL` fallbacking to local API.
- Forwards request headers/cookies and safe response headers including `Set-Cookie`, content type, cache control, and request ID.
- Streams backend response body.
- Returns `BACKEND_OFFLINE` JSON response on backend fetch failure.

## Form patterns observed

### `apps/web`

- `apps/web/src/components/auth/login-form.tsx` uses React Hook Form plus `zodResolver` and API schema from `~/lib/api/auth`.
- Auth form submits through `loginAction` in `apps/web/src/app/actions/auth.ts`.
- Form handles loading state, field errors, toast errors, permission fetch, auth store, and redirect.

### `apps/client`

- `apps/client/app/features/shared/crud-form-dialog.tsx` uses local values/errors state and `zod` schema `safeParse`.
- Shared CRUD dialog supports text/textarea/select/switch/number-style fields.
- Submit passes parsed data to caller and caller handles errors.

## Hard rules

- Both `apps/web` and `apps/client` are active; check both when backend contract changes.
- Do not duplicate API helpers or UI primitives across apps without reason.
- Do not treat frontend route hiding or UI state as authorization.
- `apps/client` lint script is placeholder-only; use typecheck/build/E2E as appropriate.

## Tests and evidence paths

- `apps/web/src/app/api/v1/[...path]/route.ts`
- `apps/client/app/routes/api-proxy.ts`
- `apps/web/src/lib/api/*`
- `apps/client/app/lib/api/*`
- `apps/web/src/components/auth/login-form.tsx`
- `apps/client/app/features/shared/crud-form-dialog.tsx`
