# Go Service Workflow

## Purpose

Primary workflow for backend module logic changes in Go: usecases, repositories, module constructors, shared backend packages, and dependency wiring.

## Use When

- changing module usecase logic
- changing repository behavior
- updating module constructor wiring
- editing shared backend package that supports module behavior
- changing worker-owned backend behavior without frontend ownership as primary concern

## Do Not Use When

- change is mostly route contract; pair or prefer `api-endpoint.md`
- change is mostly schema delta; pair or prefer `database-migration.md`
- change is cross-stack producer-consumer sync; pair or prefer `cross-stack-change.md`

## Required Read Order

1. `AGENTS.md`
2. `llm/cache/backend-map.md`
3. `llm/cache/module-map.md`
4. relevant domain cache files
5. `internal/config/app.go`
6. target module under `internal/modules/*`
7. `llm/workflows/benchmarking.md` before refactor/improve/performance-sensitive logic changes

## Live Code to Inspect

- target `module.go`
- target controller/usecase/repository/entity/model
- related middleware when auth/tenant/Casbin context involved
- shared packages like `pkg/querybuilder`, `pkg/tx`, `pkg/storage`, `pkg/tus`, `pkg/ws`, `pkg/sse`

## Workflow Steps

### Step 1 — Find True Owner Layer

Ask first:

- is logic in usecase, repository, or shared package?
- is constructor wiring affected?
- is there worker or middleware coupling?

### Step 2 — Trace Context and Dependency Flow

For nontrivial changes, trace:

- request context propagation
- transaction context propagation
- organization context propagation
- enforcer/task distributor/storage dependencies

### Step 3 — Patch Smallest Correct Owner

Before patching existing hot path, run benchmark audit from `llm/workflows/benchmarking.md`.

If no relevant benchmark exists, record that gap before refactor and decide whether to add a small benchmark seam.

Preferred order:

- shared infra only if truly owner
- repository for persistence-specific behavior
- usecase for business behavior
- constructor for wiring only if dependency graph changed

### Step 4 — Check Boundary Coupling

Re-check whether change also affects:

- auth/session
- tenant/org scope
- Casbin/policy behavior
- worker side effects
- upload/realtime/shared packages

If yes, mention it explicitly in verification and final handoff.

## Common Mistakes

- patching controller when usecase is owner
- bypassing tx/context propagation
- editing global app config when module-local fix is enough
- forgetting shared package owner under `pkg/*`

## Verification

### Minimum

- narrow package tests for touched module/package

### Add When Needed

- integration tests when DB/Redis/Casbin/tenant/worker behavior changed
- `pnpm go:test`
- `make test-integration` for boundary-crossing logic

## Stop Conditions

Stop and mark `needs confirmation` if:

- ownership unclear between module and shared package
- context/transaction semantics cannot be proven
- behavior is actually route-contract or cross-stack issue in disguise

## Completion Output

Report:

- true owner layer
- dependencies touched
- context/transaction implications
- verification run
