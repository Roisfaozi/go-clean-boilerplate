# Benchmarking Workflow

## Purpose

Benchmark-first workflow for refactor, codebase improve, logic optimization, and performance-sensitive changes in this repo.

This workflow exists because the current benchmark surface is not final: `make bench` exists, CI has a benchmark job, and pprof exists, but no real `BenchmarkXxx` functions are currently present in live Go packages.

## Use When

Use before:

- refactoring existing Go logic
- improving codebase maintainability where performance could change
- optimizing query/filter/sort behavior
- changing auth, tenant, Casbin, API-key, upload, realtime, worker, or audit logic
- changing repository queries or transaction boundaries
- changing frontend proxy/runtime behavior that may affect latency

Use after the change too, when meaningful, to compare before/after behavior.

## Do Not Use When

- task is docs-only and no runtime behavior changes
- pure UI copy/visual adjustment has no runtime path impact
- change is trivial and no measurable hot path exists

If unsure, do a short benchmark audit and record whether no valid benchmark exists.

## Required Read Order

1. `AGENTS.md`
2. `llm/conventions/testing.md`
3. relevant domain cache file
4. target module/package tests
5. `Makefile` benchmark targets
6. `.github/workflows/ci.yml` benchmark job
7. `cmd/api/main.go` and pprof config if profiling runtime server behavior

## Current Repo Benchmark Truth

Confirmed surfaces:

- `make bench` exists and runs `go test -run=^$ -bench=. -benchmem -benchtime=$(BENCHTIME) -count=5 ./...`
- `make bench-race` exists for race-enabled benchmark sanity, not stable timing comparison
- CI has a `benchmark` job that runs `make bench`
- pprof is available behind `PPROF_ENABLED` and `PPROF_PORT`

Confirmed gap:

- no real Go `BenchmarkXxx` functions were found under `internal`, `pkg`, `tests`, or `cmd` during audit
- therefore current benchmark command is infrastructure placeholder unless new benchmark cases are added

## Benchmark Decision Matrix

### Existing Benchmark Exists

Run baseline before change:

- `make bench`
- or targeted `go test -run=^$ -bench=<Pattern> -benchmem -count=5 ./pkg/or/module`

Save output if comparing:

- before: `/tmp/casbin-bench-before.txt`
- after: `/tmp/casbin-bench-after.txt`

### No Benchmark Exists, Hot Path Changed

Add or propose a small benchmark before refactor when the hot path is measurable.

Good candidates:

- `pkg/querybuilder`
- `internal/middleware` auth/API-key/tenant/Casbin paths
- `internal/modules/permission` batch checks and policy operations
- `internal/modules/audit` dynamic query/outbox path
- `pkg/ws` broadcast/presence fanout
- `pkg/tus` metadata binding and hook dispatch
- `internal/worker` enqueue/process overhead

### No Meaningful Benchmark Exists And Task Is Not Perf-Sensitive

Record in final output:

- benchmark audit performed
- no existing benchmark found for changed path
- no benchmark added because change is not performance-sensitive

Do not pretend `make bench` proves performance if it only ran without benchmark cases.

## Workflow Steps

### Step 1 — Identify Hot Path

Before editing code, state whether target path is:

- CPU-bound
- allocation-heavy
- DB/query-bound
- Redis/network-bound
- middleware route path
- async worker path
- frontend proxy/browser path

### Step 2 — Check Existing Benchmark Coverage

Search for relevant benchmark cases:

- `rg -n "func Benchmark|testing\.B|b \*testing\.B" internal pkg tests cmd -g '*.go'`

If none, record gap before changing code.

### Step 3 — Capture Baseline

If a benchmark exists or is added as pre-refactor evidence:

- run targeted benchmark first
- save output when comparison matters
- keep environment notes, especially Docker/Snap/CI differences

### Step 4 — Implement Change

Only after baseline decision:

- refactor or optimize the target code
- keep scope tight
- avoid unrelated cleanup that pollutes benchmark comparison

### Step 5 — Re-run Benchmark / Verification

After change:

- rerun same benchmark command
- compare with same `BENCHTIME`, `-count`, package path, and environment
- report allocation changes as well as ns/op

If `benchstat` is available, use it for before/after comparison. If not, report raw before/after output and say `benchstat` was not run.

### Step 6 — Pair With Correct Tests

Benchmark does not replace correctness tests.

Run tests appropriate to changed path:

- package tests for local logic
- integration tests for DB/Redis/Casbin/tenant/worker behavior
- E2E/manual flow for route or frontend-visible behavior

## Commands

Repo-level current command:

```bash
make bench
```

Targeted package example:

```bash
go test -run=^$ -bench=. -benchmem -benchtime=1s -count=5 ./pkg/querybuilder
```

Before/after capture example:

```bash
go test -run=^$ -bench=. -benchmem -benchtime=1s -count=5 ./pkg/querybuilder > /tmp/casbin-bench-before.txt
# change code
go test -run=^$ -bench=. -benchmem -benchtime=1s -count=5 ./pkg/querybuilder > /tmp/casbin-bench-after.txt
benchstat /tmp/casbin-bench-before.txt /tmp/casbin-bench-after.txt
```

pprof runtime diagnostics example:

```bash
PPROF_ENABLED=true PPROF_PORT=6060 make run
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
```

## Common Mistakes

- running `make bench` and claiming performance confidence when no `BenchmarkXxx` exists
- using benchmark result as correctness proof
- comparing before/after with different package paths or `BENCHTIME`
- benchmarking race detector output as normal performance number
- mixing large unrelated refactor with benchmark comparison
- ignoring allocations/op when refactor claims efficiency

## Stop Conditions

Stop and mark `needs confirmation` if:

- user asks for performance claim but no measurable path exists
- benchmark requires production-like data that is unavailable locally
- change affects DB/Redis/Casbin and only microbenchmark exists
- benchmark would require mutating user-owned uncommitted work

## Completion Output

Report:

- whether benchmark audit ran
- existing benchmark found or missing
- command used before change
- command used after change
- before/after result or reason benchmark was not run
- correctness tests run separately
- residual performance risk
