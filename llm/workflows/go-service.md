# Go Service Workflow

## Use when

- implementing or changing backend business logic inside Go modules
- changing service/usecase/repository wiring in `internal/modules/*`
- changing worker-triggered domain behavior owned by backend modules

## Read first

- `llm/cache/backend-map.md`
- `llm/cache/module-map.md`
- `llm/cache/domain-rules.md`
- `llm/conventions/golang.md`
- `llm/conventions/testing.md`

## Live code to inspect

- `internal/config/app.go` for dependency wiring
- target `internal/modules/*/module.go`
- target controller/usecase/repository/model files
- `pkg/transaction`, `pkg/querybuilder`, `pkg/storage`, `pkg/tus`, or other shared backend package if touched
- related tests near owning package

## Steps

1. find owning usecase and dependency graph.
2. confirm whether change is pure business logic, transaction flow, storage flow, or side effect orchestration.
3. preserve context propagation and existing module boundaries.
4. keep cross-module wiring changes in constructors/app config, not ad hoc globals.
5. add or adjust tests near changed package when pattern exists.
6. run narrow backend verification first.

## Verification commands

- narrow package tests with `go test ./path/...` or repo equivalent
- broader backend sweep when needed: `pnpm go:test`
- side effects involving DB/Redis/worker/auth/tenant: `pnpm go:test-integration`

## Review checklist

- usecase owns business rules, not controller
- repository owns persistence detail, not usecase call sites
- transactions wrap all-or-nothing writes correctly
- tenant/session/Casbin constraints not bypassed
- constructor changes do not pass full app config where not needed

## Stop conditions / needs confirmation

- dependency ownership unclear between module and shared package
- change needs cross-module contract decision not derivable from live code
- integration-only behavior cannot be validated due missing infra
