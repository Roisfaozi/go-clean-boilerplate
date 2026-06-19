# Feature Workflow

## Purpose

Workflow ini untuk membangun capability baru di repo Casbin tanpa kehilangan owner boundary, auth and tenant rules, atau consumer sync.

Ini bukan checklist generik. Workflow harus memandu agent dari owner discovery sampai verification dan handoff.

## Use when

- menambah feature backend baru
- menambah feature frontend baru di `apps/web` atau `apps/client`
- memperluas module yang sudah ada dengan behavior baru yang user-visible
- ada kemungkinan backend, frontend, worker, auth, tenant, API-key, upload, atau contract berubah bersama

## Read first

1. `AGENTS.md`
2. `llm/cache/project-overview.md`
3. `llm/cache/architecture.md`
4. `llm/cache/domain-rules.md`
5. `llm/cache/module-map.md`
6. `llm/cache/frontend-map.md` jika frontend relevan
7. `llm/cache/api-contracts.md` jika API boundary relevan
8. relevant `llm/conventions/*.md`
9. workflow lebih spesifik jika task ternyata API-only, DB-only, frontend-redesign, atau cross-stack

## Live code to inspect

- `internal/config/app.go`
- `internal/router/router.go`
- target module under `internal/modules/*`
- target frontend app under `apps/web` or `apps/client`
- shared packages under `packages/*` if reused by feature

## Workflow phases

### Phase 1 — Find owner

Tentukan:

- owner module backend
- owner app frontend
- apakah feature ini pure backend, pure frontend, atau cross-stack
- apakah ada async side effect, upload, realtime, atau auth-sensitive behavior

Jika owner belum jelas, berhenti dan gunakan `brainstorm` atau `design` dulu.

### Phase 2 — Define smallest coherent slice

Sebelum implementasi, state:

- user flow atau operator flow
- persisted behavior yang berubah
- request/response contract yang berubah
- boundary yang disentuh: auth, tenant, Casbin, API-key, worker, upload, realtime

Slice pertama harus sekecil mungkin tapi tetap end-to-end masuk akal.

### Phase 3 — Decide implementation order

Urutan default:

1. backend contract and domain behavior
2. shared types or proxy updates
3. frontend consumption
4. broader verification and review

Jangan mulai dari UI dulu kalau persisted behavior atau API contract belum jelas.

### Phase 4 — Use narrower workflows when needed

Load workflow atau skill yang lebih spesifik bila applicable:

- API route: `llm/workflows/api-endpoint.md`
- backend module logic: `llm/workflows/go-service.md`
- cross-stack: `llm/workflows/cross-stack-change.md`
- DB/schema: `llm/workflows/database-migration.md`
- frontend design or redesign: `llm/workflows/frontend-design.md`, `llm/workflows/frontend-redesign.md`, `llm/workflows/image-to-frontend.md`

### Phase 5 — Verification

Start narrow:

- package or module tests
- app typecheck or build for touched frontend
- route-level or consumer-level checks when contract moved

Escalate to integration or E2E when route, auth, tenant, worker, upload, or cookie behavior changed.

### Phase 6 — Record active state

Untuk pekerjaan multi-step:

- use `llm/tasks/todo.md` for active execution state
- use `llm/plans/` for durable phased plan
- use `llm/research/` for evidence-heavy design or comparison

## Review checklist

- owner layer benar
- auth, tenant, API-key, and Casbin boundaries preserved
- request and response contract still match consumers
- shared package reuse preferred over duplication
- no unrelated files changed

## Stop conditions / needs confirmation

- route ownership unclear between modules
- contract source of truth conflicts between backend and frontend
- feature implies migration or destructive data change without explicit user request
- env or runtime dependency is critical but not verifiable locally
