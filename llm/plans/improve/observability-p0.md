# Observability P0 Plan - Correctness and Trust

## Objective

Make the existing observability surfaces truthful and configurable before
adding more signals. P0 addresses confirmed defects that can disable tracing,
misreport health, or display fabricated metrics.

## Scope Rules

- Execute slices in order.
- Write a failing targeted test before each code fix where a stable test seam
  exists.
- Keep each slice independently reviewable.
- Do not introduce OTEL sampling, worker metrics, or broad logging cleanup in
  P0; those belong to P1/P2.
- Treat realtime payload changes as cross-stack contracts.
- Do not modify unrelated worktree changes.

## Baseline Gate

Before P0 execution:

```bash
git status --short
go test ./internal/config ./internal/middleware ./internal/router ./internal/modules/stats/... ./pkg/telemetry
pnpm --filter casbin-web typecheck
pnpm --filter casbin-client typecheck
docker compose -f docker-compose.dev.yml config
docker compose -f docker-compose.prod.yml config
```

Record pre-existing failures in `llm/tasks/observability-progress.md`.

## P0-1 - Make OTEL Configuration Load Reliably

### Problem

`Telemetry` uses inactive `env:` tags and has no Viper defaults or explicit
assignments. `OTEL_*` variables can therefore be silently dropped.

### Files

- `internal/config/config.go`
- `internal/config/config_test.go`
- `.env.example`
- `README.md`
- `documentation/guides/OBSERVABILITY.md`

### Planned Change

1. Add Viper defaults:

```go
v.SetDefault("otel.enabled", false)
v.SetDefault("otel.service_name", defaultTelemetryServiceName)
v.SetDefault("otel.collector_url", defaultTelemetryCollectorURL)
```

2. Explicitly assign after `Unmarshal`, near the Metrics assignments:

```go
cfg.Telemetry.Enabled = v.GetBool("otel.enabled")
cfg.Telemetry.ServiceName = v.GetString("otel.service_name")
cfg.Telemetry.CollectorURL = v.GetString("otel.collector_url")
```

3. Add explicit `mapstructure` tags or named config structs for Metrics and
   Telemetry. Remove inactive `env:`/`envDefault:` tags so they no longer
   advertise incorrect behavior.
4. Standardize documented metrics credentials to `METRICS_USERNAME` and
   `METRICS_PASSWORD`, matching effective Viper keys.
5. Add tests for defaults and environment overrides.

### Failing-First Tests

- `TestNewConfig_TelemetryDefaults`
- `TestNewConfig_TelemetryEnvBinding`
- `TestNewConfig_MetricsCredentialsEnvBinding`

The telemetry override test must fail before the fix by observing false/empty
telemetry values.

### Acceptance Criteria

- `OTEL_ENABLED=true` produces `cfg.Telemetry.Enabled == true`.
- Service name and collector URL use documented defaults when unset.
- Environment overrides are honored.
- Metrics credential variable names are consistent across code, tests, and
  documentation.
- Invalid or empty required values fail with actionable errors when telemetry
  is enabled.

### Verification

```bash
go test ./internal/config -run 'Telemetry|MetricsCredentials' -count=1 -v
go test ./internal/config ./pkg/telemetry
```

## P0-2 - Correct Production Health and Environment Wiring

### Problem

Production health checks call `/health`, but the router exposes
`/api/v1/health`. Production Compose uses `APP_ENV`, while the application
reads `SERVER_APP_ENV`.

### Files

- `docker-compose.prod.yml`
- `.env.example`
- deployment documentation if it describes production variables

### Planned Change

1. Change both blue and green health checks to:

```text
http://localhost:8080/api/v1/health
```

2. Replace `APP_ENV=production` with `SERVER_APP_ENV=production`.
3. Correct database variable naming if Compose uses `MYSQL_DATABASE` while the
   application reads `MYSQL_DBNAME`; retain container-specific MySQL variables
   only on the database service.
4. Add explicit operational variables to both application replicas:
   - `LOG_LEVEL`
   - `METRICS_ENABLED`
   - `METRICS_AUTH_ENABLED`
   - `METRICS_USERNAME`
   - `METRICS_PASSWORD`
   - `OTEL_ENABLED`
   - `OTEL_SERVICE_NAME`
   - `OTEL_COLLECTOR_URL`
   - `SERVER_TRUSTED_PROXIES`
   - `CORS_ALLOWED_ORIGINS`
5. Do not add Prometheus/Jaeger services to production Compose in this slice.
   Production may use external systems; P0 only makes injection explicit.

### Simulation

Render Compose and inspect the final environment for both replicas. Health
checks should resolve to the live route. No container startup is required for
the static gate, but a real health probe is required when Docker dependencies
are available.

### Acceptance Criteria

- Both application replicas receive `SERVER_APP_ENV=production`.
- Health checks use `/api/v1/health`.
- Observability variables have one documented effective name.
- `docker compose config` succeeds without undefined observability variables
  when a complete env file is supplied.

### Verification

```bash
docker compose --env-file .env.example -f docker-compose.prod.yml config
```

Optional runtime verification:

```bash
docker compose -f docker-compose.prod.yml up -d
docker compose -f docker-compose.prod.yml ps
```

## P0-3 - Remove Fabricated System Metrics

### Recommended Decision

Remove fields that cannot currently be measured truthfully. Do not replace
hard-coded values with another approximation.

### Problem

`GetSystemInsights` returns fixed latency, error rate, uptime, and active role.
The WebSocket broadcaster also sends fixed CPU, memory, and thread values.
Both frontend applications display these values as if they are operational
measurements.

### Files

- `internal/modules/stats/model/stats_model.go`
- `internal/modules/stats/usecase/stats_usecase.go`
- `internal/modules/stats/test/stats_usecase_test.go`
- `internal/config/app_helpers.go`
- `apps/web/src/lib/api/stats.ts`
- `apps/web/src/app/[locale]/dashboard/_components/system-insights.tsx`
- `apps/client/app/stores/realtime-store.ts`
- `apps/client/app/pages/DashboardPage.tsx`
- relevant frontend tests or fixtures

### Planned Contract

Realtime payload retains only values currently backed by runtime state:

```json
{
  "type": "metrics_update",
  "data": {
    "rps": 2.5,
    "active_users": 4,
    "total_users": 120,
    "scope": "instance"
  }
}
```

`scope: "instance"` makes process-local RPS and active WebSocket connections
explicit. `total_users` remains database-backed.

System insights endpoint options:

1. Recommended: temporarily remove the endpoint and frontend panel if it has
   no real field left.
2. Alternative: keep only a real `most_active_role`, but implement the query
   before returning it. Do not retain hard-coded `role:admin`.

The implementation must select one option before editing routes. Removing a
route requires API docs and consumer updates in the same slice.

### Failing-First Tests

- Backend stats test must assert that no fixed latency/error/uptime values are
  returned.
- Broadcaster payload test must assert exact allowed keys and `scope`.
- Frontend tests must assert graceful display before the first realtime event.

### Acceptance Criteria

- No user-visible operational field is populated by a literal placeholder.
- Both frontend applications compile against the new payload.
- Dashboard copy distinguishes instance-local values from cluster totals.
- Removed fields are absent from shared types, fixtures, and displays.

### Verification

```bash
go test ./internal/modules/stats/... ./internal/config -count=1
pnpm --filter casbin-web typecheck
pnpm --filter casbin-web test
pnpm --filter casbin-client typecheck
```

## P0-4 - Normalize Error-Rate Semantics

### Problem

One frontend treats `0.02` as `0.02%`; the other treats it as a ratio and
displays `2.0%`.

### Resolution

If P0-3 removes error rate, delete all remaining error-rate fields and display
logic. This is the recommended outcome.

If the project instead chooses to preserve error rate, define the wire value as
a ratio in `[0,1]`, rename it `error_rate_ratio`, and multiply by 100 only in
presentation code. Add boundary tests for `0`, `0.02`, and `1`.

### Acceptance Criteria

- A single unit is documented and enforced.
- Both frontends display identical results from identical payloads.
- No `||` fallback changes a valid zero into a placeholder value; use nullish
  coalescing where appropriate.

## P0-5 - P0 Review and Documentation Closure

### Tasks

1. Run the complete P0 verification matrix.
2. Inspect `git diff` for unrelated edits and accidental generated churn.
3. Update `llm/tasks/observability-progress.md` slice by slice.
4. Update `llm/research/observability-audit-2026-08.md` only if live facts have
   changed and are committed/stable.
5. Regenerate Swagger only if an endpoint or response contract changed.
6. Record residual risks that move to P1.

### P0 Exit Gate

P0 is complete only when:

- telemetry environment configuration is covered by tests;
- production health and env mappings render correctly;
- no fabricated system metric remains in API/WebSocket/frontend contracts;
- backend and both active frontend surfaces pass their relevant checks;
- skipped Docker/browser checks have an exact blocker recorded.
