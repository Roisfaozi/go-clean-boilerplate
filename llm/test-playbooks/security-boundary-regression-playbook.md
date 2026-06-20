# Security Boundary Regression Playbook

## Purpose

Playbook ringkas untuk verifikasi setelah perubahan auth, tenant, Casbin, API key, upload, worker, realtime, atau frontend contract.

## Change-Type Checklist

### Auth / Session

- verify login, refresh, logout, revoke-all, reset-password
- verify Redis-backed session truth
- verify `authenticated` routes still reject API-key-only requests

### Tenant / Organization / Casbin

- verify `RequireOrganization()` and `OptionalOrganization()` semantics
- verify member cache invalidation after invite/update/remove/restore/delete
- verify Casbin domain is `organization_id` when required and `global` when intended

### API Key

- verify route scopes on create/list/delete flows
- verify wildcard scope behavior only where intended
- verify `RequireUserSession()` blocks API-key-only access for user-session routes

### Worker / Audit / Webhook

- verify task enqueue ordering relative to DB transaction
- verify retry/idempotency on handlers
- verify worker side effects stay observable

### Upload / TUS

- verify auth on upload route
- verify metadata trust boundary
- verify completion hook cannot mutate wrong user/org

### Realtime / SSE / WS

- verify ticket validation for WS
- verify SSE auth and presence updates
- verify unsubscribe/cleanup on disconnect

### Frontend Contract

- verify backend route change against `apps/web` proxy
- verify backend route change against `apps/client` proxy
- run app typecheck/build when response shape changes

## Narrow Verification Priority

1. package tests near changed code
2. router or middleware tests for boundary changes
3. integration tests for DB/Redis/Casbin/worker changes
4. E2E only when route or consumer behavior is user-visible
5. race tests for WS/worker/cache stateful paths

## Known Environment Blockers

- Snap Go can block `go test` and `make test-race` in this environment.
- Docker/Testcontainers required for integration/E2E.
- `apps/client` lint is placeholder-only and should not be treated as meaningful verification.

## Reporting Rules

- state exactly what passed
- state exactly what failed
- state exact blocker when verification cannot run
- do not claim race safety unless race detector actually ran
