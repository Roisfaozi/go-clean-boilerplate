# Casbin Repo Deep Audit

## Scope

Audit ini rangkum kelemahan dan kekurangan repo dari live code, fokus ke fitur, arsitektur, testing, dan risiko operasional.

## Pertanyaan Audit

- dimana coupling paling tinggi?
- boundary keamanan mana paling rawan?
- fitur mana paling kompleks dan rentan regresi?
- test coverage mana kuat dan mana lemah?

## Evidence Paths

- `cmd/api/main.go`
- `internal/config/config.go`
- `internal/config/app.go`
- `internal/router/router.go`
- `internal/middleware/auth_middleware.go`
- `internal/middleware/api_key_middleware.go`
- `internal/middleware/tenant_middleware.go`
- `internal/middleware/casbin_middleware.go`
- `internal/modules/auth/module.go`
- `internal/modules/organization/module.go`
- `internal/modules/api_key/module.go`
- `internal/modules/user/module.go`
- `internal/modules/webhook/module.go`
- `internal/worker/processor.go`
- `pkg/tus/handler.go`
- `pkg/ws/ws_manager.go`
- `tests/README.md`
- `package.json`
- `apps/web/package.json`
- `apps/client/package.json`

## Verified Facts

- Startup root gabung banyak concern di satu composition root: DB, Redis, JWT, Casbin, WS, SSE, TUS, storage, worker, scheduler, telemetry.
- Route strata kompleks: public, authenticated, tenantAuthorized, admin, upload.
- Authenticated route pakai kombinasi API key, JWT, auto-scope, require-user-session, user-status, tenant, Casbin.
- Tenant resolution jadi prasyarat authZ domain.
- Casbin middleware fail-open saat enforcer `nil` di non-release mode.
- Frontend aktif dua surface: `apps/web` dan `apps/client`.
- `apps/client` lint masih placeholder.
- Test suite luas dan dibagi unit/integration/e2e; integration pakai singleton Docker containers.

## Inference

- Sistem ini dirancang untuk feature breadth, bukan simplicity.
- Security model bergantung order middleware dan correctness context propagation.
- Regression risk tinggi di auth, tenant, API key, worker side effect, upload metadata, realtime, dan query/filter.
- Test breadth tinggi, tapi quality gate frontend client masih tidak sekuat backend.

## Weaknesses by Area

### Architecture

- `internal/config/app.go` terlalu gemuk; satu file bangun dependency graph besar dan sulit diuji terpisah.
- `internal/router/router.go` memuat banyak policy orchestration; perubahan kecil route bisa ubah security boundary.
- Startup `cmd/api/main.go` campur service lifecycle dan scheduler startup; partial failure handling berisiko membingungkan.

### Auth, Tenant, Casbin

- Auth correctness tidak cukup JWT; Redis session, ticket WS, API key scope, status user, dan org context ikut menentukan akses.
- `internal/middleware/casbin_middleware.go` fail-open di env non-release saat enforcer nil.
- Tenant route bergantung resolusi organisasi dari header/context/param; salah resolusi berarti domain authorization salah.

### Upload, Realtime, Worker

- TUS upload trust boundary tetap besar; router hanya outer auth, sedangkan metadata/hook logic tetap kritis.
- WS presence dan Redis pub/sub punya stateful behavior yang sulit diverifikasi tanpa E2E atau race-focused testing.
- Worker/outbox/webhook async side effects membuat request success ≠ side effect completed.

### Query/Data

- Query/filter safety harus fail-closed; penambahan field cepat bisa bocor data sensitif.
- DB-backed Casbin dan transactional writes menambah risk sinkronisasi policy vs data.

### Frontend

- Dua frontend aktif berarti API contract drift bisa terjadi jika proxy/type layer tidak diuji bersama backend.
- `apps/client` lint placeholder menurunkan signal quality gate.

### Testing

- Unit + integration + e2e sudah ada, tapi infra-heavy suite bergantung Docker/Testcontainers.
- Race test lulus dari output user, jadi isu utama bukan race yang mudah direproduksi.
- Coverage kuat pada middleware/usecase, tapi belum ada bukti penuh untuk contract drift lintas frontend/backend di semua surface.

## Priority Risks

1. middleware order dan context propagation
2. API key scope plus user-session distinction
3. tenant/org resolution dan Casbin domain
4. TUS metadata/hook trust boundary
5. async worker side effects
6. querybuilder field allowlist
7. dual-frontend contract drift

## Recommendations

- pecah audit lanjutan per module: auth, organization, api_key, permission, user, webhook, worker, tus, ws, stats, project, access, role
- tambah contract matrix antara router, middleware, dan module usecase untuk route sensitif
- audit path upload/realtime/worker dengan E2E/race lens, bukan unit saja
- rawat frontend client quality gate; jangan biarkan lint placeholder jadi satu-satunya barometer

## Deep Module Audit

### Auth

- Owner: `internal/modules/auth`.
- Coupling: token repo Redis, user repo, org repo, transaction manager, event publisher, Casbin adapter, task distributor, WebSocket ticket manager, SSO providers.
- Strength: auth module sudah menggabungkan session storage, audit/event side effect, SSO, and ticket issuance di usecase boundary.
- Weakness: dependency surface sangat lebar; auth change berpotensi menyentuh Redis, DB, Casbin, worker, SSO, WS, SSE.
- Risk: logout/revocation, SSO callback, registration default role, and session concurrency harus dites full lifecycle.
- Verification need: middleware tests plus auth integration/e2e, bukan unit saja.

### Organization and Tenant

- Owner: `internal/modules/organization` plus `internal/middleware/tenant_middleware.go`.
- Coupling: org repo, member repo, invitation repo, user repo, Redis cache, transaction manager, permission enforcer, task distributor.
- Strength: tenant is explicit boundary and route groups require organization context where needed.
- Weakness: org can be resolved from header, context, route id, or slug; ambiguity here can produce wrong Casbin domain.
- Risk: stale membership cache after invite/member updates/restore/delete, superadmin deleted-org access, and cross-org data leakage.
- Verification need: organization integration, tenant isolation E2E, cache invalidation tests.

### API Key

- Owner: `internal/modules/api_key` plus `internal/middleware/api_key_middleware.go`.
- Coupling: user repo, Redis, org context, scope derivation, route-level explicit scopes.
- Strength: route groups distinguish API-key identity from user-session-only paths with `RequireUserSession()`.
- Weakness: scope auto-derivation plus explicit scopes can drift when routes change.
- Risk: wildcard scopes, project/org scoping, and API-key-only access to user-session routes.
- Verification need: API-key middleware tests plus lifecycle integration/e2e for protected route families.

### Permission and Casbin

- Owner: `internal/modules/permission`, `internal/middleware/casbin_middleware.go`, Casbin config.
- Coupling: role repo, user repo, access repo, audit usecase, transactional enforcer behavior.
- Strength: authorization uses subject, domain, object, method rather than role-only checks.
- Weakness: fail-open Casbin behavior in non-release mode can hide policy config mistakes in dev/test.
- Risk: policy writes out of sync with DB writes, route path mismatch, wildcard policy overreach.
- Verification need: permission integration, route matrix tests, transactional enforcer regression tests.

### Role and Access

- Owner: `internal/modules/role`, `internal/modules/access`.
- Coupling: role usecase calls permission usecase; access rights feed permission expansion.
- Strength: roles and access rights are separate modules, which helps semantics stay explicit.
- Weakness: access-right registry drift can break permission expansion silently.
- Risk: role delete/update policy cleanup, access soft delete, role-access uniqueness migration assumptions.
- Verification need: role/access integration plus permission assignment checks.

### User

- Owner: `internal/modules/user`.
- Coupling: transaction manager, enforcer, audit, auth, webhook, storage.
- Strength: user lifecycle can coordinate auth revocation, audit, webhook, and avatar storage.
- Weakness: broad side effects make user usecase a high-blast-radius module.
- Risk: delete/suspend status handling, avatar mutation via direct API and TUS hook, search/list field exposure.
- Verification need: user lifecycle integration, querybuilder tests for list/search, upload completion tests for avatar.

### Project

- Owner: `internal/modules/project`.
- Coupling: minimal module wiring, but route protection depends on tenantAuthorized group and API-key scopes.
- Strength: simpler module boundary than auth/user/org.
- Weakness: security is mostly externalized to router/middleware, so route registration mistakes matter.
- Risk: project list/get/update/delete without correct org context or scope.
- Verification need: project integration/e2e with tenant isolation and API key scope.

### Audit, Webhook, Worker

- Owner: `internal/modules/audit`, `internal/modules/webhook`, `internal/worker`.
- Coupling: audit usecase, outbox, Asynq distributor/processor, webhook task handlers, email/cleanup jobs.
- Strength: async side effects avoid blocking request path and have dedicated worker handlers.
- Weakness: async semantics are harder to reason about; success response can precede side-effect completion.
- Risk: enqueue before commit, retry idempotency, webhook delivery logging, partial failures hidden from caller.
- Verification need: worker unit tests, integration tests around request plus outbox/worker side effect, retry/idempotency checks.

### Upload and Storage

- Owner: `pkg/tus`, `pkg/storage`, user avatar hook path.
- Coupling: auth context, metadata, storage driver, completion hook registry, user update path.
- Strength: TUS is separated from CRUD controllers and storage is abstracted.
- Weakness: upload metadata is a trust boundary and not obvious from router alone.
- Risk: spoofed metadata, wrong hook type, storage URL shape differences, orphaned files if domain update fails.
- Verification need: `pkg/tus` tests, storage tests, E2E upload completion scenarios.

### Realtime

- Owner: `pkg/ws`, `pkg/sse`, realtime routes.
- Coupling: Redis pub/sub, presence manager, WS tickets, organization-aware channels, frontend hooks.
- Strength: WebSocket uses ticket validation instead of raw token at upgrade.
- Weakness: distributed broadcast and presence state are concurrency/stateful surfaces.
- Risk: stale presence, duplicate broadcast, missing unsubscribe cleanup, frontend channel contract drift.
- Verification need: WS/SSE tests, race tests, E2E realtime flows, frontend hook contract checks.

### Stats

- Owner: `internal/modules/stats`.
- Coupling: direct DB-backed usecase, authenticated routes.
- Strength: simple module wiring.
- Weakness: aggregation correctness and tenant scoping depend on query implementation.
- Risk: cross-tenant aggregate leakage or stale realtime metric assumptions.
- Verification need: stats integration and tenant-isolated assertions.

### Frontend Contract Surfaces

- Owner: `apps/web`, `apps/client`, `packages/*`.
- Strength: shared workspace packages exist for api types/hooks/utils/ui.
- Weakness: two active apps require duplicated contract awareness.
- Risk: `apps/web/src/proxy.ts` only does cookie gate for dashboard pages and does not replace backend auth; `apps/client/app/routes/api-proxy.ts` forwards request path and selected response headers, so backend cookie/header behavior must be checked end to end.
- Verification need: backend route test plus both frontend typecheck/build/proxy or browser checks when API shape changes.

## Deep Codebase Gaps

- No single route authorization matrix document is generated from router + middleware; reviewers must manually inspect `internal/router/router.go`.
- No automatic guarantee that frontend proxy paths, shared API types, and backend route contracts stay aligned.
- Composition root makes dependency graph visible but not modular; startup and module wiring tests can become brittle.
- Testcontainers dependency is operationally heavy; failures can be infra-related rather than product-related.
- Race detector output from user passed, but cached packages and infra logs mean targeted race tests should still be run when changing WS/worker/cache code.

## Suggested Next Audit Artifacts

- `llm/research/auth-tenant-casbin-route-matrix.md`
- `llm/research/worker-audit-webhook-failure-modes.md`
- `llm/research/upload-realtime-contract-audit.md`
- `llm/research/frontend-contract-drift-audit.md`

## Needs Confirmation

- apakah semua env release benar-benar selalu menyalakan Casbin enforcer
- apakah semua frontend proxy path masih sinkron dengan backend routes terbaru
- apakah seluruh async worker side effect sudah idempotent saat retry
