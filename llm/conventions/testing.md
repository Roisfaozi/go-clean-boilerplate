# Testing Conventions

## Purpose

Guide for choosing the right verification layer, understanding infra assumptions, and reporting validation honestly in this repo.

## Test layers

- unit
  - package-local tests under `internal` and `pkg`
- integration
  - `tests/integration/...`
- E2E
  - `tests/e2e/...`
- frontend client E2E
  - `apps/client/tests/e2e`

## Typical evidence by layer

- unit
  - handler, usecase, repository, or package tests
- integration
  - real MySQL or Redis interactions under `tests/integration`
- E2E
  - route lifecycle and security scenarios under `tests/e2e`
- frontend E2E
  - login or RBAC browser tests under `apps/client/tests/e2e`

## Core commands

- unit: `make test` or `make test-unit`
- coverage: `make test-coverage`
- integration: `make test-integration`
- E2E: `make test-e2e`
- all backend tests: `make test-all`
- bench: `make bench`
- benchmark workflow: `llm/workflows/benchmarking.md`
- frontend client E2E: package script `test:e2e`
- frontend web checks: app scripts `lint`, `typecheck`, `build`
- frontend client checks: `lint`, `typecheck`, `build`, `test:e2e`; `lint` uses Biome and `typecheck` remains separate
- security-sensitive change playbooks: `llm/test-playbooks/security-boundary-change-types.md`

## Infrastructure assumptions

- integration and E2E rely on Docker
- worker lifecycle matters in integration/E2E when async side effects are part of behavior
- Redis and MySQL are not optional for broad integration behavior
- restricted environments can block Snap Go or Docker; report exact blocker and run narrowest useful fallback

## Change-type routing

### Middleware, auth, session

- package tests near `internal/middleware`
- auth integration or E2E if route behavior changed

### Tenant, organization, Casbin

- tenant middleware tests
- permission or organization integration tests
- tenant isolation E2E when route behavior is user-visible

### API key

- API-key middleware or usecase tests
- lifecycle integration or E2E when route behavior changed

### Upload or TUS

- `pkg/tus` tests
- storage package tests
- integration or E2E when upload completion affects domain behavior

### Worker, audit, webhook, email

- `internal/worker` tests
- integration where request semantics depend on async side effects

### Query builder

- `pkg/querybuilder` tests
- dynamic search tests for affected repositories or modules

### Frontend proxy or API boundary

- backend route verification
- `apps/web` or `apps/client` typecheck, build, or E2E as relevant

## Strategy rules

- prefer package-level tests when change is internal logic
- prefer integration tests when change affects DB, Redis, Casbin, worker, or stateful runtime
- prefer E2E when route-group, cookie, tenant, or full lifecycle behavior is user-visible
- prefer failing-first bug reproduction when seam is meaningful
- before refactor, improve, optimization, or logic rewrite, perform benchmark audit first
- if no relevant `BenchmarkXxx` exists, report that `make bench` is placeholder-grade for that path unless benchmark cases are added

## Benchmarking rules

- use `llm/workflows/benchmarking.md` before codebase improve/refactor/performance-sensitive logic changes
- benchmark before and after when a meaningful benchmark exists
- do not claim performance improvement from `make bench` if no benchmark functions cover the changed path
- benchmark results never replace correctness tests
- keep before/after commands identical: same package, same `BENCHTIME`, same `-count`, same environment
- prefer targeted package benchmarks over repo-wide benchmark runs when investigating one hotspot

## Mock and fixture rules

- regenerate mocks with `make mocks` when interfaces change
- keep unit tests isolated with mocks or local test doubles
- use integration containers for DB/Redis behavior instead of over-mocking persistence semantics
- keep auth or tenant fixtures close to scenario tests so route-group differences stay readable

## Reporting rules

- separate passed checks from skipped checks
- if Docker is unavailable, say integration/E2E were not run because Docker is required
- if frontend lint is run, say it uses Biome and does not replace `typecheck`
- do not imply success from unrun checks
