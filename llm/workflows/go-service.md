# Go Service Workflow

## Purpose

Workflow ini untuk perubahan backend business logic di Go modules, termasuk usecase, repository, constructor wiring, transaction-sensitive behavior, dan backend-owned side effects.

## Use when

- mengubah logic di `internal/modules/*`
- mengubah dependency wiring service/usecase/repository
- mengubah side effect backend-owned yang bukan sekadar route registration
- mengubah business rules tanpa menjadikan route/API contract sebagai concern utama

## Read first

1. `AGENTS.md`
2. `llm/cache/backend-map.md`
3. `llm/cache/module-map.md`
4. `llm/cache/domain-rules.md`
5. relevant domain cache files
6. `llm/conventions/golang.md`
7. `llm/conventions/testing.md`

## Live code to inspect

- `internal/config/app.go` untuk dependency wiring
- target `internal/modules/*/module.go`
- target controller, usecase, repository, model files
- `pkg/tx`, `pkg/querybuilder`, `pkg/storage`, `pkg/tus`, `pkg/ws`, `pkg/sse` jika disentuh
- related tests near owning package

## Workflow phases

### Phase 1 — Find owning usecase

Tentukan apakah perubahan utamanya ada di:

- usecase business rules
- repository behavior
- transaction flow
- async side effect orchestration
- constructor wiring

### Phase 2 — Preserve boundaries

Pastikan:

- controller hanya bind/validate/respond
- usecase owns business rules
- repository owns persistence details
- constructor or app wiring owns dependency composition

### Phase 3 — Trace side effects and shared packages

Jika perubahan menyentuh:

- transactions
- storage
- querybuilder
- worker or webhook side effects
- auth or tenant semantics

maka trace package boundary itu dulu sebelum patch.

### Phase 4 — Patch minimal owner layer

- preserve context propagation
- avoid ad hoc globals or bypassed dependency injection
- keep cross-module calls explicit and justified

### Phase 5 — Verification

Start with narrow package tests.

Escalate to broader backend or integration checks when DB, Redis, worker, auth, tenant, or upload semantics move.

## Review checklist

- usecase owns business rules
- repository owns persistence details
- transaction boundaries still all-or-nothing where required
- tenant or session or Casbin constraints not bypassed
- constructor changes do not pass full app config where not needed

## Stop conditions / needs confirmation

- dependency ownership unclear between module and shared package
- change requires cross-module contract decision not derivable from live code
- integration-only behavior cannot be validated due missing infra
