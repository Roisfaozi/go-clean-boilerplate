# Feature Workflow

## Purpose

Primary workflow for building new capability in this Casbin repo without losing owner boundaries, auth/tenant rules, consumer sync, or verification honesty.

This is top-level orchestration workflow. Use narrower workflows underneath it when scope becomes more specific.

## Use When

- adding new backend feature
- adding new frontend feature in `apps/web` or `apps/client`
- extending existing module with new user-visible behavior
- feature may touch backend, frontend, worker, auth, tenant, API-key, upload, realtime, or shared contract

## Do Not Use When

- task is clearly only bugfix, route patch, schema patch, or redesign and already fits narrower workflow exactly

## Required Read Order

1. `AGENTS.md`
2. `llm/cache/project-overview.md`
3. `llm/cache/architecture.md`
4. `llm/cache/domain-rules.md`
5. `llm/cache/module-map.md`
6. `llm/cache/frontend-map.md` if frontend relevant
7. `llm/cache/api-contracts.md` if API boundary relevant
8. relevant `llm/conventions/*.md`
9. `llm/workflows/benchmarking.md` if implementation changes existing hot path or refactors logic
10. narrower workflow once feature shape known

## Workflow Steps

### Step 1 — Find Owners and Boundaries

State:

- owner backend module
- owner frontend app if any
- whether feature is backend-only, frontend-only, or cross-stack
- whether async side effect, upload, realtime, auth, tenant, or API-key boundary is involved

If owner unclear, stop and design/brainstorm first.

### Step 2 — Define Smallest Coherent Slice

Before implementation, define:

- user or operator flow
- persisted behavior that changes
- route/API contract that changes
- boundary risks touched

First slice must be as small as possible while still end-to-end meaningful.

### Step 3 — Choose Narrower Workflow

Use the more specific workflow when feature shape is known:

- API route: `llm/workflows/api-endpoint.md`
- backend module logic: `llm/workflows/go-service.md`
- cross-stack: `llm/workflows/cross-stack-change.md`
- DB/schema: `llm/workflows/database-migration.md`
- frontend code/design: `frontend-design.md`, `frontend-redesign.md`, `image-to-frontend.md`, `image-generation.md`

### Step 4 — Implement In Stable Order

Default order:

1. backend/domain behavior
2. shared types / proxies / contracts
3. frontend consumption
4. async or secondary integrations
5. verification and handoff

Do not start from UI if persisted behavior and route contract are still unclear.

### Step 4.5 — Benchmark Before Refactor Or Improve

If feature requires refactoring existing code, changing query behavior, rewriting core logic, or touching auth/tenant/Casbin/upload/realtime/worker hot paths:

- run `llm/workflows/benchmarking.md` before implementation
- capture existing benchmark if present
- record benchmark gap if no `BenchmarkXxx` covers target path
- do not claim performance impact without before/after evidence

### Step 5 — Record Active State Correctly

For multi-step work:

- `llm/tasks/` for active scratch/task state
- `llm/plans/` for durable phased implementation plan
- `llm/research/` for evidence-heavy design and comparison notes
- `llm/recommendations/` for deferred follow-up not in current scope

## Common Mistakes

- starting implementation before owner discovery
- letting feature creep swallow unrelated cleanup
- ignoring auth/tenant/API-key implications because feature feels “small”
- changing one consumer while another active app remains stale
- writing cache facts before feature behavior is stable/committed

## Verification

### Minimum

- narrowest package/app checks covering changed slice

### Add When Needed

- integration when route, DB, Redis, Casbin, tenant, worker, upload, or cookie behavior changed
- E2E/manual browser validation for protected or multi-step user flows
- docs generation when public contract changed

## Stop Conditions

Stop and mark `needs confirmation` if:

- route ownership unclear between modules
- contract source of truth conflicts between backend and frontend
- feature implies migration or destructive data change not requested
- critical env/runtime dependency cannot be verified locally and assumption would be risky

## Completion Output

Report:

- feature slice delivered
- owners touched
- narrower workflows used
- verification run
- deferred next slice if feature is intentionally staged
