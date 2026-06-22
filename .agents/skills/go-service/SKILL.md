---
name: go-service
description: Use when changing Go backend module logic, usecases, repositories, module constructors, worker-owned backend behavior, or dependency wiring in the Casbin backend.
---

# Go Service: Backend Module Change

**Announce at start:** "I'm using the go-service skill to preserve Casbin repo backend boundaries."

## When To Use

| Use this skill when...                     | Use another skill when...                                         |
| ------------------------------------------ | ----------------------------------------------------------------- |
| changing usecase/repository/module wiring  | route protection changes -> `api-endpoint` + `auth-tenant-casbin` |
| changing business rules                    | schema migrations -> `database-transactions`                      |
| changing worker-triggered backend behavior | frontend contract changes -> `cross-stack-change`                 |

## Read Order

1. `AGENTS.md`
2. `llm/cache/backend-map.md`
3. `llm/cache/module-map.md`
4. `llm/cache/domain-rules.md`
5. `llm/conventions/golang.md`
6. `llm/workflows/go-service.md`
7. `internal/config/app.go` when lifecycle/constructor changes
8. target `internal/modules/*/module.go`
9. target controller/usecase/repository/model/tests

## Backend Boundary Map

- controller: bind/validate/request-response only
- usecase: business rules, transaction orchestration, side-effect decisions
- repository: GORM/query details only
- `internal/config/app.go`: dependency graph and constructor wiring
- `internal/router/router.go`: HTTP route/middleware placement

## Workflow

### Phase 1 — Locate Owner

- Identify owning module and exact usecase method.
- Check constructor dependencies in `module.go` and app wiring.
- Identify cross-module calls and side effects: audit, webhook, worker, storage, Casbin, Redis.

### Phase 2 — Risk Classification

Mark touched boundaries:

- auth/session
- tenant/org membership/cache
- Casbin policy/enforcer
- API key
- TUS/storage
- worker/audit/webhook
- querybuilder/list filtering
- realtime/SSE/WebSocket

Load matching project-specific skill if any boundary is touched.

### Phase 3 — Patch Rules

- Do not pass full app config into usecases.
- Preserve context propagation.
- Keep transactional writes together.
- Use transactional enforcer patterns when Casbin policy writes share DB semantics.
- Do not move business rules into handlers.
- Do not make tenant checks optional.

### Phase 4 — Test Strategy

- Use package-local unit tests for pure usecase/repository logic.
- Use integration tests when DB/Redis/Casbin/worker/tenant boundaries are involved.
- Use E2E tests when route lifecycle/cookies/tokens/frontend proxy behavior changes.

## Review Checklist

- [ ] dependency ownership clear
- [ ] no full config object passed into usecase
- [ ] context propagation intact
- [ ] repository hides GORM details
- [ ] side effects intentionally sync/async
- [ ] tests or code trace cover changed behavior

## Stop Conditions

- Stop and ask before destructive DB/schema/data operations not explicitly requested.
- Stop if live code contradicts `llm/cache/*`; live code wins, then document drift in `llm/tasks/`.
- Stop if route ownership, tenant boundary, or auth stratum is unclear.

## Completion Output

Report:

- files changed
- commands run and exact result
- verification skipped and exact blocker
- risks or follow-up work
