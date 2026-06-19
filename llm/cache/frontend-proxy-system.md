# Frontend Proxy and Forms System

## Purpose

Durable map for active frontend API proxy surfaces, shared client boundaries, and reusable form behavior patterns.

## Active frontend surfaces

- `apps/web`: Next.js App Router, server actions, UI components, backend API proxy
- `apps/client`: React Router 7, feature folders, route registry, backend API proxy
- `packages/*`: shared types, hooks, UI, and utils

## API proxy surfaces

### `apps/web`

- Proxy file: `apps/web/src/app/api/v1/[...path]/route.ts`
- Builds backend URL from `NEXT_PUBLIC_API_URL`, falling back to local API URL
- Reads `access_token` cookie and sets `Authorization: Bearer ...` when present
- Forwards cookies and selected safe response headers
- Returns `BACKEND_OFFLINE` JSON on backend fetch failure

### `apps/client`

- Proxy file: `apps/client/app/routes/api-proxy.ts`
- Builds backend base URL from `NEXT_PUBLIC_API_URL`, falling back to local API URL
- Forwards request headers and cookies plus safe response headers including `Set-Cookie`
- Streams backend response body back to callers
- Returns `BACKEND_OFFLINE` JSON response on backend fetch failure

## Contract and auth implications

- Both active apps can depend on backend contract changes through their own proxy boundary.
- Cookie and auth forwarding behavior is part of runtime correctness, not a frontend implementation detail.
- If backend contract changes, check producer and consumer, not backend alone.

## Form patterns observed

### `apps/web`

- `apps/web/src/components/auth/login-form.tsx` uses React Hook Form plus `zodResolver`
- auth form submits through `apps/web/src/app/actions/auth.ts`
- flow handles loading state, field errors, toast errors, permission fetch, auth store, and redirect

### `apps/client`

- `apps/client/app/features/shared/crud-form-dialog.tsx` uses local values/errors state and `zod` schema `safeParse`
- shared CRUD dialog supports text, textarea, select, switch, and number-style fields
- submit passes parsed data to caller and caller handles result or error behavior

## Hard rules

- Both `apps/web` and `apps/client` are active; check both when backend contract changes.
- Do not duplicate API helpers or proxy behavior across apps without reason.
- Do not treat frontend route hiding or UI state as authorization.
- `apps/client` lint script is placeholder-only; use typecheck, build, or E2E as appropriate.
- If form payload or response shape changes, audit both validation and proxy path.

## Verification and evidence paths

- `apps/web/src/app/api/v1/[...path]/route.ts`
- `apps/client/app/routes/api-proxy.ts`
- `apps/web/src/lib/api/*`
- `apps/client/app/lib/api/*`
- `apps/web/src/components/auth/login-form.tsx`
- `apps/client/app/features/shared/crud-form-dialog.tsx`
