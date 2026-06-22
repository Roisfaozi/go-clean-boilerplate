# Frontend Contract Drift Audit

## Scope

Audit ini fokus ke drift antara backend route contract dan dua frontend surface: `apps/web` dan `apps/client`.

## Evidence Paths

- `internal/router/router.go`
- `apps/web/src/proxy.ts`
- `apps/client/app/routes/api-proxy.ts`
- `apps/web/package.json`
- `apps/client/package.json`
- `packages/api-types/*`
- `packages/hooks/*`
- `packages/ui/*`

## Verified Facts

- `apps/web` dan `apps/client` sama-sama aktif.
- `apps/web` proxy melakukan auth gate di browser layer untuk dashboard route, tapi bukan pengganti backend auth.
- `apps/client` proxy forward request ke backend base URL dan meneruskan body/header terpilih.
- Shared workspace packages exist for API types, hooks, and UI.
- `apps/client` lint masih placeholder.

## Weaknesses

- Two active frontend apps increase contract drift risk.
- Backend route changes can silently break one app while app lain masih jalan.
- Proxy behavior is runtime contract, not just plumbing.
- Without route matrix, frontend ownership mapping stays manual.

## High-Risk Contract Areas

1. auth/session cookies and bearer flow
2. tenant-scoped endpoints and org headers/params
3. API key routes and scope expectations
4. upload completion and avatar mutation
5. realtime hooks and ticket flow
6. audit export and stats dashboards

## Recommendations

- Maintain backend endpoint -> shared type -> app consumer map.
- Add route change checklist that requires checking both proxies.
- Replace placeholder client lint with meaningful validation.
- Tie contract changes to typecheck/build checks per app.

## Needs Confirmation

- exact list of shared API types consumed by each app
- whether all backend routes are proxied through both apps or some are app-specific
- whether frontend ownership should be documented in repo-native matrix
