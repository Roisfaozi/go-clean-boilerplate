# Testing Conventions

## Layers

- Unit: package-local tests under `internal` and `pkg`.
- Integration: `tests/integration/...`.
- E2E: `tests/e2e/...`.
- Frontend E2E for client: `apps/client/tests/e2e`.

Typical evidence by layer:

- unit: handler/usecase/repository package tests under `internal` and `pkg`
- integration: real MySQL/Redis tests under `tests/integration`
- E2E: route lifecycle and security scenarios under `tests/e2e`
- frontend E2E: login/RBAC browser tests under `apps/client/tests/e2e`

## Commands

- Unit: `make test` or `make test-unit`.
- Coverage: `make test-coverage`.
- Integration: `make test-integration`.
- E2E: `make test-e2e`.
- All backend tests: `make test-all`.
- Benchmarks: `make bench`.
- Frontend client E2E: package script `test:e2e`.
- Frontend web checks: app scripts `lint`, `typecheck`, `build`.
- Frontend client checks: app scripts `typecheck`, `build`, `test:e2e`; `lint` is placeholder-only.

## Infrastructure assumptions

- Integration and E2E rely on Docker.
- Project docs define singleton-container pattern for integration testing.
- Worker lifecycle matters in E2E/integration when async side effects are part of behavior.
- Redis and MySQL are not optional for broad integration behavior.
- Snap-packaged Go may fail in restricted environments; if that happens, report exact blocker and run the narrowest available check.

Current repo testing signals:

- CI runs lint, unit, integration, E2E, benchmark, and build checks.
- frontend client has Playwright E2E already wired.
- backend has wide regression coverage around auth, tenant, rate limit, realtime, worker, and security scenarios.

## What to run by change type

- Middleware/auth/session: `internal/middleware` tests plus auth integration/E2E if route behavior changes.
- Tenant/org/Casbin: tenant middleware tests, permission/organization integration, tenant isolation E2E.
- API key: API key middleware/usecase tests plus API key lifecycle integration/E2E.
- Upload/TUS: `pkg/tus` tests plus TUS integration/E2E.
- Worker/audit/webhook/email: `internal/worker` tests plus worker integration scenarios.
- Query builder: `pkg/querybuilder` tests and dynamic search tests for affected modules.
- Frontend proxy/API boundary: backend route tests plus `apps/web` or `apps/client` type/build/E2E as relevant.

Strategy note:

- prefer package-level tests when the change is internal logic
- prefer integration tests when the change affects DB/Redis/Casbin/worker/stateful runtime
- prefer E2E when route-group, cookie, or full lifecycle behavior is user-visible

## Mock and fixture conventions

- Regenerate mocks with `make mocks` when interfaces change.
- Keep unit tests isolated with mocks or local test doubles.
- Use integration containers for DB/Redis behavior instead of over-mocking persistence semantics.
- Clean test data between integration tests rather than restarting containers when following singleton-container pattern.

Fixture tip:

- keep route/auth fixtures close to scenario tests so tenant and session setup stays readable
- avoid overusing shared fixtures that hide route-group differences

## Reporting validation

- Separate passed checks from skipped checks.
- If Docker is unavailable, say integration/E2E were not run because Docker is required.
- If frontend `apps/client` lint is run, note it is placeholder-only and not a quality signal.
