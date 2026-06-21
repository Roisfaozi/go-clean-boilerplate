# Security Boundary Hardening Plan

## Scope

Perbaikan bertahap berdasarkan audit kelemahan repo, fokus pada auth, tenant, Casbin, API key, cache invalidation, async worker, upload, realtime, frontend contract, dan testing.

## Source Audit

- `llm/research/casbin-deep-audit.md`
- `llm/research/auth-tenant-casbin-route-matrix.md`
- `llm/research/route-protection-matrix.md`

## Current Progress

- Phase 0 baseline started with route group semantics captured.
- Phase 1 completed with route matrix artifact in `llm/research/route-protection-matrix.md`.
- Representative middleware tests added for API-key vs user-session and Casbin path normalization.
- Phase 2 completed: auth session methods moved to `internal/modules/auth/usecase/auth_session.go`; recovery, ticket, and SSO methods moved to `internal/modules/auth/usecase/auth_recovery_sso.go`; host-run `go test ./internal/modules/auth/... -v` passes.
- Phase 3 completed at unit level: organization member invite/update/remove/accept-invitation and organization delete cache invalidation assertions pass on host-run `go test ./internal/modules/organization/... -v`.
- Phase 4 baseline verification slice continued: representative API-key/JWT chain tests, Casbin trailing-slash test, extracted auto-scope derivation tests, and strict non-local Casbin startup guard are in place; miniredis-backed API-key package tests still need host/socket-capable verification.
- Regression playbook created in `llm/test-playbooks/security-boundary-regression-playbook.md`.
- Phase 5 research artifact created in `llm/research/webhook-worker-failure-modes.md`.
- Phase 5 implementation complete at code/testable package level: webhook handler surfaces webhook-log persistence failures for retryable delivery failures without converting successful or permanent 4xx deliveries into retries, audit outbox sync uses the outbox ID as the audit-log ID to prevent duplicate audit rows on replay, socket-free worker handler/tasks plus audit/webhook package tests pass, and Redis-backed worker root tests still require host/socket-capable verification.
- Phase 6 completed: avatar completion rejects client-controlled legacy `user_id` metadata and requires server-bound `authenticated_user_id`; TUS completion hook failures now terminate completed storage when the tusd store supports termination; storage, TUS/user, WebSocket ticket/presence, and SSE tests pass with socket-dependent checks verified from host shell.
- Phase 7 research artifact created in `llm/research/frontend-contract-map.md`.

## Non-Goals

- tidak mengubah product behavior tanpa test kontrak
- tidak menambah migration kecuali ada persetujuan eksplisit
- tidak mengganti framework frontend/backend
- tidak memindahkan authorization truth ke frontend

## Execution Rules

- runtime truth wajib dari live code
- setiap phase harus kecil dan reviewable
- auth/tenant/Casbin/API-key change wajib punya test route-level atau integration sesuai boundary
- frontend contract change wajib cek `apps/web` dan `apps/client`
- worker/upload/realtime change wajib cek async/idempotency/stateful behavior

## Phase 0 — Baseline and Safety Net

### Objective

Buat baseline sebelum refactor/hardening supaya perubahan berikutnya bisa diverifikasi.

### Owners

- `llm/research/casbin-deep-audit.md`
- `llm/research/auth-tenant-casbin-route-matrix.md`
- `internal/router/router.go`
- `internal/middleware/*`
- `tests/README.md`

### Tasks

1. Capture route group semantics from `internal/router/router.go`.
2. List representative endpoint per group: public, authenticated, tenantAuthorized, authorized, upload, realtime.
3. Identify current passing tests for each representative endpoint.
4. Mark missing tests as explicit gaps.

### Verification

- no code behavior change expected
- run doc-only sanity: paths referenced exist
- optional: `git diff --check`

### Gate

Proceed only after route groups and representative endpoints are agreed.

## Phase 1 — Route Protection Matrix

### Objective

Turn implicit middleware order into durable matrix.

### Owners

- `llm/research/auth-tenant-casbin-route-matrix.md`
- new or updated route matrix doc under `llm/research/` or `llm/test-playbooks/`
- `internal/router/router.go`
- route registration files under `internal/modules/*/delivery/http/*routes*.go`

### Tasks

1. Generate table columns:
   - route
   - HTTP method
   - router group
   - JWT required
   - API key allowed
   - user session required
   - org required
   - Casbin domain
   - explicit API-key scopes
   - test coverage
2. Verify route rows against live router and module route files.
3. Add review note for route placement rules:
   - use `authenticated` for user-session-only routes
   - use `tenantAuthorized` for org domain routes
   - use `authorized` for admin/global routes
   - use upload group only for TUS transport

### Verification

- static verification via `rg` against router and route files
- no runtime test required unless code generation script is added

### Gate

Matrix must be complete for all protected route families before modifying middleware.

## Phase 2 — Auth Boundary Split Plan

### Objective

Reduce `AuthUseCase` blast radius without changing behavior first.

### Owners

- `internal/modules/auth/usecase/auth_usecase.go`
- `internal/modules/auth/usecase/interfaces.go`
- `internal/modules/auth/module.go`
- `internal/modules/auth/repository/token_repository.go`
- `internal/modules/auth/test/*`

### Tasks

1. Identify internal seams:
   - session lifecycle: generate, verify, revoke, revoke all, refresh
   - credential auth: login, lockout, password check
   - registration provisioning: user create, default role, default org
   - recovery: forgot/reset/verify email
   - SSO session issuance
   - WebSocket ticket issuance
2. Add tests around current behavior before extraction:
   - refresh revokes old session and stores new session
   - revoked session fails middleware verification
   - revoke-all clears session index
   - register creates user + default org + default role behavior
3. Extract only after tests cover current behavior.

### Verification

- narrow: `go test -race ./internal/modules/auth/... ./internal/middleware/...`
- broader: `make test-unit`
- integration if Redis/session semantics changed: auth integration/e2e tests

### Gate

No extraction until tests prove existing session and provisioning behavior.

## Phase 3 — Tenant Cache Hardening

### Objective

Remove stale membership/role auth risk caused by missing invalidation or negative-cache semantics.

### Owners

- `internal/modules/organization/usecase/reader.go`
- `internal/modules/organization/usecase/organization_usecase.go`
- `internal/modules/organization/usecase/organization_member_usecase.go`
- `internal/modules/organization/repository/*`
- `internal/middleware/tenant_middleware.go`
- `internal/modules/organization/test/*`
- `tests/integration/modules/organization_integration_test.go`
- `tests/e2e/api/tenant_isolation_e2e_test.go`

### Tasks

1. Inventory mutation paths:
   - create organization
   - invite member
   - accept invitation
   - update member role/status
   - remove member
   - soft delete organization
   - restore organization
   - hard delete organization
2. Verify each path invalidates:
   - membership key
   - member role key
   - organization status key when relevant
3. Review negative cache behavior:
   - consider shorter TTL for non-member entries
   - consider not caching negative entries for invite/accept-sensitive paths
4. Add tests for cache invalidation after membership mutation.

### Verification

- narrow: `go test ./internal/modules/organization/... ./internal/middleware/...`
- Redis behavior: organization integration tests
- route behavior: tenant isolation E2E

### Gate

Do not change cache TTL or negative-cache behavior without test showing expected stale-window behavior.

## Phase 4 — Casbin and API-Key Consistency

### Objective

Align API-key scope, user-session requirements, tenant domain, and Casbin policy writes.

### Owners

- `internal/middleware/api_key_middleware.go`
- `internal/middleware/casbin_middleware.go`
- `internal/modules/permission/usecase/permission_usecase.go`
- `internal/modules/role/usecase/*`
- `internal/modules/access/*`
- `internal/router/router.go`
- `tests/integration/modules/api_key_integration_test.go`
- `tests/integration/modules/permission_integration_test.go`
- `tests/integration/scenarios/transactional_enforcer_regression_test.go`

### Tasks

1. Verify all routes with API-key support have explicit or valid auto-derived scopes. ✅
2. Add tests for representative protected routes: ✅
   - JWT allowed
   - API key allowed
   - API key denied by missing scope
   - API key denied by `RequireUserSession()`
   - wildcard scope behavior
3. Audit policy writes that coincide with DB writes. ✅
4. Ensure transactional enforcer path is used where DB and policy must commit together. ✅
5. Consider stronger nil-enforcer guard for non-local environments. ✅

### Verification

- narrow: `go test ./internal/middleware/... ./internal/modules/permission/... ./internal/modules/api_key/...`
- integration: API-key + permission + transactional enforcer regression tests
- E2E: role/permission/auth route tests when route behavior changes

### Gate

No route group movement without route matrix update and focused endpoint tests.

### Phase 4 Result

- Router grouping remains explicit: `authenticated` uses `Authenticate()` + `RequireScopeAuto()` + `RequireUserSession()`, while `tenantAuthorized` uses `Authenticate()` + tenant resolution + `RequireScopeAuto()`.
- Representative route and middleware tests cover JWT allowed, API-key allowed, missing scope, `RequireUserSession()` denial, wildcard behavior, and Casbin path normalization.
- Transactional enforcer regression coverage exists for grouping policy persistence and organization-create policy persistence.
- Strict non-local Casbin startup guard is in place in `internal/config/app.go`.
- Host/Docker-backed integration proof remains environment-dependent and skips cleanly when DB/Testcontainers are unavailable.

## Phase 5 — Worker, Audit, Webhook Reliability

### Objective

Make async side effects explicit, idempotent, and observable.

### Owners

- `internal/modules/audit/*`
- `internal/modules/webhook/*`
- `internal/worker/*`
- `internal/worker/handlers/*`
- `internal/worker/tasks/*`
- `tests/integration/modules/audit_integration_test.go`
- `tests/integration/modules/webhook_integration_test.go`
- `tests/integration/modules/worker_integration_test.go`

### Tasks

1. Map request side effects that enqueue audit/webhook/email/cleanup tasks.
2. Check enqueue timing relative to DB transactions.
3. Verify worker handlers are idempotent under retry.
4. Ensure webhook delivery logs capture failures without hiding from ops.
5. Add tests for retry and duplicate task scenarios.

### Verification

- narrow: `go test -race ./internal/worker/... ./internal/modules/audit/... ./internal/modules/webhook/...`
- integration: audit/webhook/worker integration tests

### Gate

Do not change retry/idempotency semantics without integration evidence.

## Phase 6 — Upload and Realtime Hardening

### Objective

Reduce trust-boundary and stateful-concurrency risk in TUS, storage, WS, and SSE.

### Owners

- `pkg/tus/*`
- `pkg/storage/*`
- `pkg/ws/*`
- `pkg/sse/*`
- `internal/router/router.go`
- `internal/modules/user/*` for avatar hook path
- `tests/integration/modules/tus_integration_test.go`
- `tests/e2e/modules/tus_e2e_test.go`
- `tests/e2e/realtime/*`

### Tasks

1. Audit upload metadata sources and server-side overrides.
2. Whitelist hook types and reject unknown/unsafe metadata.
3. Verify avatar upload path cannot mutate another user's avatar.
4. Add cleanup strategy for storage success + domain failure.
5. Audit WS presence multi-connection behavior.
6. Audit Redis pub/sub duplicate broadcast and unsubscribe cleanup.

### Verification

- narrow: `go test -race ./pkg/tus/... ./pkg/storage/... ./pkg/ws/... ./pkg/sse/...`
- integration: TUS integration tests
- E2E: TUS and realtime E2E tests

### Gate

No upload metadata change without TUS integration/e2e coverage.

## Phase 7 — Frontend Contract Drift Controls

### Objective

Keep backend route contract, shared types, and both frontend apps aligned.

### Owners

- `apps/web/src/proxy.ts`
- `apps/client/app/routes/api-proxy.ts`
- `apps/web/src/hooks/*`
- `apps/client/app/**`
- `packages/api-types/*`
- `packages/hooks/*`
- `packages/ui/*`

### Tasks

1. Build consumer map: ✅
   - backend endpoint
   - shared type
   - `apps/web` consumer/proxy
   - `apps/client` consumer/proxy
2. Add contract checklist for backend API changes. ✅
3. Replace or supplement `apps/client` placeholder lint with meaningful validation. ✅
4. Add typecheck/build expectations per app in PR checklist. ✅

### Verification

- `pnpm --filter casbin-web typecheck`
- `pnpm --filter casbin-web build` when UI/runtime boundary changes
- `pnpm --filter casbin-client typecheck`
- `pnpm --filter casbin-client build`
- `pnpm --filter casbin-client test:e2e` for browser flows

### Gate

No backend response-shape change without consumer map update.

### Phase 7 Result

- Consumer map and backend API checklist recorded in `llm/research/frontend-contract-map.md`.
- `apps/client` lint no longer returns placeholder success; it runs Biome and typecheck remains separate.
- Required frontend verification commands are documented per app/package.

## Phase 8 — Verification Pipeline and Regression Playbook

### Objective

Make verification repeatable and honest.

### Owners

- `Makefile`
- `package.json`
- `tests/README.md`
- `llm/conventions/testing.md`
- `llm/test-playbooks/*`

### Tasks

1. Create test playbooks by change type: ✅
   - auth/session
   - tenant/Casbin/API key
   - upload/storage
   - worker/webhook/audit
   - realtime
   - frontend contract
2. Document Snap Go and Docker/Testcontainers blockers. ✅
3. Add targeted race-test recommendations for stateful packages. ✅
4. Keep broad suite separate from narrow proof checks. ✅

### Verification

- doc path sanity
- optional: run targeted tests after each implemented change
- final broad checks:
  - `make test-unit`
  - `make test-integration`
  - `make test-e2e`
  - `pnpm typecheck`
  - `pnpm build`

### Phase 8 Result

- Change-type verification playbook added in `llm/test-playbooks/security-boundary-change-types.md`.
- Test playbook index and testing convention updated for current frontend/client lint behavior.
- Security regression playbook now includes race-test recommendations and environment blockers.

## Dependency Order

1. Phase 0 before all implementation.
2. Phase 1 before router/middleware changes.
3. Phase 2 can run before or after Phase 3, but not while route semantics are being changed.
4. Phase 3 before major tenant/Casbin policy changes.
5. Phase 4 after route matrix exists.
6. Phase 5 and Phase 6 can proceed independently after Phase 0.
7. Phase 7 must run whenever backend contract changes.
8. Phase 8 should be updated continuously after each phase.

## Suggested First PR Split

1. Docs-only route matrix and verification playbook.
2. Tests for auth/session and route group semantics.
3. Tenant cache invalidation tests and fixes.
4. API-key/Casbin consistency tests and fixes.
5. Worker/upload/realtime targeted hardening.
6. Frontend contract and client quality gate.

## Completion Criteria

- route protection matrix exists and matches live code
- auth/session critical flows have focused tests
- tenant membership cache invalidation has integration coverage
- API-key/Casbin boundary has representative route tests
- async worker retries/idempotency documented and tested
- upload metadata trust boundary is tested
- realtime presence behavior is tested
- both frontend apps have meaningful validation for changed contracts
