# AGENTS — Guide for automated coding agents

This document is the repository entrypoint for coding agents. Use it together with the `llm/` starter-pack files that were concretized from live code in this repo.

## 1. What this repo is

- Hybrid monorepo.
- Go backend is the operational core.
- Active frontend apps exist in `apps/web` and `apps/client`.
- Root workspace tooling uses `pnpm` and `turbo`.

Primary source-of-truth order for agents:

1. live code
2. config / entrypoint / wiring
3. tests
4. supporting docs
5. older prompts/workflows only as fallback context

## 2. Read order for agents

Start here in this order:

1. `llm/cache/project-overview.md`
2. `llm/cache/environment.md`
3. `llm/cache/architecture.md`
4. `llm/cache/backend-map.md`
5. `llm/cache/frontend-map.md` if frontend is relevant
6. `llm/cache/api-contracts.md` if API/frontend boundary is relevant
7. `llm/cache/module-map.md`
8. `llm/cache/domain-rules.md`
9. relevant files under `llm/conventions/`
10. relevant files under `llm/workflows/`

Then verify against live code before changing anything.

Fast routing by task type:

- backend feature: read `llm/workflows/go-service.md`, `llm/cache/backend-map.md`, `llm/cache/module-map.md`.
- benchmarking/refactor/improve before code change: read `llm/workflows/benchmarking.md`, `llm/conventions/testing.md`, then target package tests.
- API route change: read `llm/workflows/api-endpoint.md`, `llm/cache/api-contracts.md`, `internal/router/router.go`.
- frontend change: read `llm/cache/frontend-map.md`, `llm/cache/frontend-proxy-system.md`, `llm/conventions/typescript.md`, then the target app package.
- frontend design/new UI flow: read `llm/workflows/frontend-design.md`, `llm/references/frontend-skill-map.md`, then target app package.
- frontend redesign flow: read `llm/workflows/frontend-redesign.md`, `llm/references/frontend-skill-map.md`, then target existing app surface.
- image-led frontend implementation: read `llm/workflows/image-to-frontend.md`, `llm/references/frontend-skill-map.md`, then target app package.
- cross-stack change: read `llm/workflows/cross-stack-change.md`, `llm/cache/api-contracts.md`, `llm/cache/frontend-proxy-system.md`, both frontend proxy files.
- DB/schema change: read `llm/workflows/database-migration.md`, `llm/conventions/database.md`, `db/migrations`.
- auth/session change: read `llm/cache/authentication-system.md`, `llm/cache/domain-rules.md`, `internal/middleware/auth_middleware.go`.
- tenant/organization change: read `llm/cache/tenant-organization-system.md`, `llm/cache/user-system.md`, `internal/middleware/tenant_middleware.go`, target organization usecase.
- Casbin/permission change: read `llm/cache/casbin-permission-system.md`, `llm/cache/permission-system.md`, `llm/cache/role-system.md`, `llm/cache/access-right-system.md`, `internal/middleware/casbin_middleware.go`, `internal/modules/permission`.
- API-key change: read `llm/cache/api-key-system.md`, `internal/middleware/api_key_middleware.go`, `internal/modules/api_key`.
- upload/storage change: read `llm/cache/tus-upload-system.md`, `pkg/tus`, `pkg/storage`.
- worker/audit/webhook change: read `llm/cache/worker-audit-webhook-system.md`, `llm/cache/audit-system.md`, `llm/cache/webhook-system.md`, `internal/worker`, target audit/webhook module.
- query/filter/sort change: read `llm/cache/querybuilder-security.md`, `llm/cache/user-system.md`, `pkg/querybuilder`.
- realtime change: read `llm/cache/realtime-system.md`, `llm/cache/stats-system.md`, `pkg/ws`, `pkg/sse`, `internal/router/router.go`.
- user/profile/avatar/list change: read `llm/cache/user-system.md`, `llm/cache/querybuilder-security.md`, `pkg/tus`, `internal/modules/user`.
- role/permission change: read `llm/cache/role-system.md`, `llm/cache/access-right-system.md`, `llm/cache/permission-system.md`, `llm/cache/casbin-permission-system.md`, `internal/modules/role`, `internal/modules/permission`.
- project change: read `llm/cache/project-system.md`, `llm/cache/api-key-system.md`, `llm/cache/tenant-organization-system.md`, `internal/modules/project`, `internal/router/router.go`.
- stats change: read `llm/cache/stats-system.md`, `internal/modules/stats`, `internal/config/app.go`.
- security/auth/tenant/Casbin change: read `llm/cache/domain-rules.md`, the matching domain cache, `internal/middleware/*`, target usecase.

## 3. Core runtime truth

Primary runtime files:

- `cmd/api/main.go`
- `internal/config/app.go`
- `internal/router/router.go`
- `internal/config/config.go`

These files define startup, dependency wiring, route strata, middleware layering, env mapping, and service composition.

## 4. Main repository areas

- `cmd/api/` — API binary entrypoint.
- `cmd/gen/` — module scaffolding generator.
- `internal/config/` — app wiring, config, DB, Redis, Casbin, storage.
- `internal/modules/*` — business modules using repository/usecase/controller boundaries.
- `internal/middleware/` — auth, tenant, API key, Casbin, rate-limit, logging, recovery, metrics.
- `internal/worker/` — Asynq distributor, processor, scheduler, task handlers.
- `pkg/` — shared infra/helpers such as JWT, SSE, WS, TUS, storage, tx, query builder.
- `db/migrations/` — SQL schema migrations.
- `db/seeds/` — seed scripts.
- `tests/` — integration and E2E test suites.
- `apps/web/` — Next.js frontend.
- `apps/client/` — React Router frontend.
- `packages/*` — shared workspace packages used by frontend apps.

## 5. Hard repo rules

- Do not pass full app config into usecases; pass only required values/dependencies.
- Keep context propagation intact, especially around storage and request-scoped operations.
- Casbin changes that share DB transaction semantics must use transactional enforcer patterns.
- Treat organization/tenant handling as a first-class boundary, not an optional add-on.
- Keep auth/session checks in middleware/usecase boundaries, not duplicated ad hoc in handlers.
- Do not weaken query-builder field restrictions for sensitive fields.
- Do not move route protection from router/middleware into frontend-only checks.
- Do not assume JWT validity is enough; Redis-backed session validation matters.
- Do not bypass API-key scope checks when adding protected endpoints.
- Do not add migrations without paired up/down SQL files.

High-risk areas that need extra review:

- auth token/cookie/session changes
- tenant organization resolution and membership cache changes
- Casbin policy writes and route matcher changes
- API key auth/scope behavior
- TUS upload metadata and hook dispatch
- WebSocket origin/ticket/presence behavior
- query builder field allow/deny behavior
- worker side effects that affect audit, webhook, email, or cleanup behavior

## 6. Confirmed command surfaces

Root workspace:

- `pnpm build`
- `pnpm lint`
- `pnpm test`
- `pnpm typecheck`
- `pnpm go:run`
- `pnpm go:test`
- `pnpm go:test-integration`
- `pnpm go:test-e2e`
- `pnpm go:docs`

Backend / Makefile:

- `make run`
- `make build`
- `make test`
- `make test-unit`
- `make test-integration`
- `make test-e2e`
- `make test-all`
- `make test-coverage`
- `make bench`
- `make docs`
- `make mocks`
- `make docker-dev`
- `make migrate-up`
- `make migrate-down`
- `make seed-up`
- `make gen-module`

Frontend packages:

- `apps/web`: `dev`, `build`, `start`, `lint`, `typecheck`, `format`
- `apps/client`: `dev`, `build`, `start`, `typecheck`, `test:e2e`

Note:

- `apps/client` lint script is currently placeholder-only and should not be treated as strong verification.

## 7. Testing expectations

- Unit tests live close to packages under `internal` and `pkg`.
- Integration tests live under `tests/integration` and require Docker.
- E2E tests live under `tests/e2e` and require Docker.
- Frontend client has Playwright E2E under `apps/client/tests/e2e`.
- Regenerate mocks with `make mocks` when interfaces change.

Verification strategy:

- Start with the narrowest package or app check that covers the change.
- Use integration tests when a change crosses DB, Redis, Casbin, worker, tenant, or upload boundaries.
- Use E2E tests when route behavior, cookies/tokens, frontend proxies, or full request lifecycle changes.
- If Docker or Snap Go blocks validation locally, report the exact blocker and the best narrower check that still ran.

## 8. Workflow routing

Use these workflow files instead of improvising:

- `llm/workflows/feature.md`
- `llm/workflows/bugfix.md`
- `llm/workflows/benchmarking.md`
- `llm/workflows/api-endpoint.md`
- `llm/workflows/go-service.md`
- `llm/workflows/cross-stack-change.md`
- `llm/workflows/database-migration.md`
- `llm/workflows/frontend-design.md`
- `llm/workflows/frontend-redesign.md`
- `llm/workflows/image-to-frontend.md`

Frontend skill routing note:

- use `llm/references/frontend-skill-map.md` to avoid loading conflicting frontend taste/style/reference skills all at once.

## 9. Knowledge lifecycle folders

Use each folder intentionally:

- `llm/cache/` — verified stable repo facts only; do not put active plans or uncommitted assumptions here.
- `llm/tasks/` — active work state, phase tracking, parity audits, lessons, and current task notes.
- `llm/research/` — durable investigations with evidence and explicit separation between facts and recommendations.
- `llm/recommendations/` — non-urgent improvement ideas backed by evidence.
- `llm/test-playbooks/` — reusable manual/API/browser/E2E verification flows.
- `llm/plans/` — large staged plans that should outlive one active task; use `improve/` and `roadmap/` for durable plan categories.

Memory/update rule:

- only move information into `llm/cache/` after it is verified against committed live code or committed docs generated from live code.
- if a finding is useful but not yet stable, keep it in `llm/tasks/`, `llm/research/`, or `llm/recommendations/` instead.

## 10. Agent behavior for this repo

Before making claims in docs or patches:

- verify commands against `package.json`, `Makefile`, package scripts, or CI
- verify paths exist in filesystem
- verify route/auth/tenant behavior against `internal/router/router.go` and middleware/usecase code
- prefer live code over README when they differ

When task touches architecture-heavy areas:

- inspect `internal/config/app.go` first
- inspect `internal/router/router.go` second
- inspect target module constructor and usecase/repository/controller flow third

When editing docs or AI context:

- keep claims concrete and tied to live paths
- prefer `needs confirmation` only when the fact cannot be verified locally
- update `llm/tasks/lessons.md` when a durable repo-specific lesson is discovered
- update `llm/tasks/phase-compliance.md` if starter-pack coverage changes
- write long-form evidence gathering into `llm/research/`, not `llm/cache/`, until findings are stable
- write non-urgent future work into `llm/recommendations/` or `llm/plans/`, not `llm/tasks/todo.md`

When editing frontend:

- remember `apps/web` and `apps/client` are both active surfaces
- check which app owns the route/component before changing shared packages
- do not treat `apps/client` lint as strong verification while its lint script is placeholder-only

## 11. If you need more context

Use the concretized starter-pack files under `llm/`. They were filled from live repo evidence and are the preferred durable handoff layer for future agent work.
