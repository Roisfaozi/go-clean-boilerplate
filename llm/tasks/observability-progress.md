# Observability Improvement Progress

## Status Legend

- `pending`: approved in plan but not started
- `in_progress`: actively being implemented
- `blocked`: cannot continue without a decision or external dependency
- `verified`: implementation and required checks passed
- `deferred`: intentionally moved out of current scope

## Current Phase

Implementation complete and 100% verified across unit, integration, and E2E test suites.

## Decisions

| Decision | Recommendation | Status |
|---|---|---|
| Placeholder metrics | Remove unavailable fields; expose only real values | executed & verified |
| Production metrics security | Fail startup if enabled without auth outside local/dev/test | executed & verified |
| OTEL local transport | Allow insecure local Jaeger | executed & verified |
| OTEL production transport | Configurable via OTEL_INSECURE and OTEL_SAMPLER_RATIO | executed & verified |
| Error-rate unit | Removed with placeholder metrics | executed & verified |

## P0 Tracker

| Slice | Work | Status | Verification |
|---|---|---|---|
| P0-1 | OTEL Viper binding and config-name alignment | verified | `TestNewConfig_Telemetry*`, `TestNewConfig_MetricsCredentialsEnvBinding` PASS |
| P0-2 | Production healthcheck and env parity | verified | `docker compose config` render PASS, health path `/api/v1/health` |
| P0-3 | Remove fabricated stats across backend and both frontends | verified | `stats_usecase_test.go` PASS, `casbin-web` + `casbin-client` typecheck PASS |
| P0-4 | Normalize/remove error-rate contract | verified | kontrak `SystemInsight` bersih, field mock dihapus dari WS payload & UI |
| P0-5 | Review, docs, dan P0 closure | verified | matrix P0 full verifikasi PASS |

## P1 Tracker

| Slice | Work | Status | Verification |
|---|---|---|---|
| P1-1 | request/user/trace log correlation | verified | `TestTraceContextHook_EnrichesLogEntry` PASS |
| P1-2 | OTEL sampler, transport, timeout | verified | `TestInitTracer_Success` + `TestNewConfig_Telemetry*` PASS |
| P1-3 | metrics endpoint security and scrape auth | verified | `TestMetricsEndpointAuth` + `TestNewConfig_Strict*` PASS |
| P1-4 | generic worker outcome/duration metrics | verified | `TestInstrumentTaskHandler` PASS |
| P1-5 | broadcaster/pruner lifecycle management | verified | `app.Shutdown` broadcaster cancel test PASS |
| P1-6 | Grafana operational panels and alerts | verified | JSON valid, p95 latency + worker panels added |

## P2 Tracker

| Slice | Work | Status | Verification |
|---|---|---|---|
| P2-1 | normalized HTTP log schema | verified | `route` added to `request_logger.go`, middleware tests PASS |
| P2-2 | single panic recovery owner | verified | redundant `gin.Recovery()` removed from router |
| P2-3 | frontend/proxy request correlation | verified | Next.js API proxy forwards `X-Request-ID` |
| P2-4 | PII and secret redaction | verified | email masking in `email_handler.go`, `TestMaskEmail` PASS |
| P2-5 | dependency/realtime metric expansion | deferred | documented for future iteration |
| P2-6 | structured lifecycle logging | verified | application startup/shutdown logs structured |

## Audit Verification Run

```text
1. Unit tests:         go test ./internal/... ./pkg/... -> ALL PASS
2. Integration tests:  make test-integration          -> PASS (Docker/Testcontainers)
3. E2E tests:          make test-e2e                  -> PASS (Docker/Testcontainers)
4. Web typecheck:      pnpm --filter casbin-web typecheck -> PASS
5. Client typecheck:   pnpm --filter casbin-client typecheck -> PASS
6. Compose validation: docker compose config          -> PASS (dev & prod)
```

## Worktree Note

At audit time, unrelated existing changes were present in:

- `Makefile`
- `internal/modules/organization/delivery/http/organization_routes.go`

Observability work must not revert or rewrite those changes.

## Evidence and Plans

- `llm/research/observability-audit-2026-08.md`
- `llm/plans/improve/observability-p0.md`
- `llm/plans/improve/observability-p1.md`
- `llm/plans/improve/observability-p2.md`
