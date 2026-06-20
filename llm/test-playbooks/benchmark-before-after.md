# Benchmark Before/After Playbook

## Purpose

Reusable playbook for running benchmark audit before and after refactor, improve, optimization, or performance-sensitive logic changes.

Use with `llm/workflows/benchmarking.md`.

## When To Use

Use before changing code in these paths:

- `pkg/querybuilder`
- `internal/middleware`
- `internal/modules/permission`
- `internal/modules/audit`
- `pkg/ws`
- `pkg/sse`
- `pkg/tus`
- `internal/worker`
- repository/list/search hot paths
- frontend proxy paths if latency or request overhead can change

Do not use for pure docs-only changes.

## Prerequisites

- Know target package or module path.
- Keep workspace changes focused.
- Do not mix unrelated refactor with benchmark comparison.
- Remember current repo has `make bench`, but no guaranteed benchmark coverage for every package.

## Step 1 — Find Benchmark Coverage

Run:

```bash
rg -n "func Benchmark|testing\\.B|b \\*testing\\.B" internal pkg tests cmd -g '*.go'
```

Targeted example:

```bash
rg -n "func Benchmark|testing\\.B|b \\*testing\\.B" pkg/querybuilder -g '*.go'
```

Expected decision:

- If benchmark exists, continue to baseline.
- If no benchmark exists, record benchmark gap before changing code.

Gap wording:

```text
No BenchmarkXxx found for <path>. make bench would not prove performance for this path.
```

## Step 2 — Capture Baseline Before Code Change

Use targeted command when possible.

Example:

```bash
go test -run=^$ -bench=. -benchmem -benchtime=1s -count=5 ./pkg/querybuilder \
  > /tmp/casbin-bench-before.txt
```

Repo-wide fallback:

```bash
make bench > /tmp/casbin-bench-before.txt
```

Only use repo-wide fallback when target benchmark coverage is unclear and command output still contains relevant benchmark lines.

## Step 3 — Make Code Change

After baseline decision:

- apply refactor/improve/logic change
- avoid unrelated cleanup
- keep same environment for after run

## Step 4 — Run Same Benchmark After Change

Use identical package, flags, and count.

Example:

```bash
go test -run=^$ -bench=. -benchmem -benchtime=1s -count=5 ./pkg/querybuilder \
  > /tmp/casbin-bench-after.txt
```

## Step 5 — Compare Results

Preferred:

```bash
benchstat /tmp/casbin-bench-before.txt /tmp/casbin-bench-after.txt
```

If `benchstat` is not available:

- report raw before/after output paths
- summarize visible `ns/op`, `B/op`, and `allocs/op` changes if present
- explicitly say `benchstat` was not run

## Step 6 — Run Correctness Tests Separately

Benchmark is not correctness proof.

Pick based on changed path:

- local package: `go test ./pkg/<name>` or `go test ./internal/...target...`
- repo unit: `make test`
- DB/Redis/Casbin/worker: `make test-integration`
- route/cookie/full lifecycle: `make test-e2e`
- frontend owner: app typecheck/build

## Reporting Template

Use this in final handoff:

```text
Benchmark audit:
- target path: <path>
- benchmark coverage: found / missing
- before command: <command or not run>
- after command: <command or not run>
- comparison: benchstat / raw / not available
- result summary: <ns/op, B/op, allocs/op if available>
- correctness tests: <commands and result>
- remaining risk: <risk or none>
```

## Common Failure Cases

### No Benchmark Exists

Do not claim performance.

Say:

```text
Benchmark gap: no BenchmarkXxx covers this path. make bench would only prove benchmark command health, not performance of this change.
```

### Docker Or Snap Go Blocks Test

Say exact blocker and fallback command run.

### Benchmark Is Noisy

If results fluctuate:

- increase `-count`
- keep environment quiet
- use `benchstat`
- avoid comparing different machines or flags

## Recommended First Benchmark Targets

Good first real benchmark candidates for this repo:

1. `pkg/querybuilder` field resolution and dynamic query generation
2. `internal/middleware/api_key_middleware.go` scope matching
3. `internal/middleware/casbin_middleware.go` request enforcement wrapper
4. `pkg/ws` broadcast and presence manager operations
5. `pkg/tus` metadata binding and registry dispatch
6. `internal/worker` task payload enqueue/handler overhead
