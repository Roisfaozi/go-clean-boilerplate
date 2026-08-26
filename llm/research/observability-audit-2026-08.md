# Observability Audit - August 2026

## Purpose

This document records the verified current state of metrics, tracing, logging,
realtime stats, and observability deployment in this repository. It is the
evidence base for the implementation plans under `llm/plans/improve/`.

No runtime change is authorized by this document. Recommendations are kept
separate from verified facts.

## Research Questions

1. How are logging, Prometheus metrics, OpenTelemetry, and application stats
   configured and wired at runtime?
2. Which environment variables are actually read by Viper?
3. Can logs currently be correlated with requests, users, and OTEL traces?
4. Are the metrics exposed to Prometheus and displayed to users real runtime
   values?
5. Do development and production deployment files support the declared
   observability behavior?
6. Which tests currently prove these behaviors?

## Primary Evidence

- `internal/config/config.go`
- `internal/config/app.go`
- `internal/config/app_helpers.go`
- `internal/config/logrus.go`
- `internal/config/gorm.go`
- `internal/router/router.go`
- `internal/middleware/request_id.go`
- `internal/middleware/request_logger.go`
- `internal/middleware/prometheus.go`
- `internal/middleware/recovery.go`
- `pkg/telemetry/tracer.go`
- `pkg/telemetry/metrics.go`
- `internal/modules/stats/usecase/stats_usecase.go`
- `internal/worker/processor.go`
- `pkg/ws/ws_manager.go`
- `docker-compose.dev.yml`
- `docker-compose.prod.yml`
- `deploy/prometheus/prometheus.yml`
- `deploy/grafana/dashboards/main_dashboard.json`
- both frontend stats and proxy consumers

## Verified Runtime Map

### Logging

- `internal/config.NewLogrus` creates one shared Logrus logger.
- Development uses colored text only when `SERVER_APP_ENV` equals exactly
  `development`; other values use JSON.
- Caller reporting is always enabled.
- `TraceContextHook` currently enriches entries with `request_id` and attempts
  to enrich them with `user_id`.
- `RequestLogger` emits HTTP method, raw URL path, status, latency, client IP,
  user agent, and response size after the handler chain completes.
- `RecoveryMiddleware` emits panic error and stack trace.
- The router installs both `gin.Recovery()` and the custom recovery middleware.
- Most infrastructure and modules receive the same logger, but not every call
  site uses `WithContext(ctx)`.
- `cmd/api/main.go` still uses the standard library logger for startup,
  scheduler lifecycle, shutdown, and fatal startup errors.

### Prometheus

- `PrometheusMiddleware` records:
  - `http_requests_total{method,path,status}`
  - `http_request_duration_seconds{method,path}`
- The `path` metric label uses `c.FullPath()`, which bounds route cardinality.
- Unknown routes use the literal label `unknown`.
- `/metrics` is registered only when `cfg.Metrics.Enabled` is true.
- `/metrics` can use Gin Basic Auth.
- Metrics are enabled by default; metrics authentication is disabled by
  default.
- `pkg/telemetry/metrics.go` defines business metrics for login, registration,
  WebSocket connections, storage uploads, and cleanup tasks.
- Go runtime/process metrics are exposed by the default Prometheus registry.

### OpenTelemetry

- `internal/config/app.go` initializes an OTLP gRPC tracer only when
  `cfg.Telemetry.Enabled` is true.
- `otelgin.Middleware` instruments Gin only when that same flag is true.
- `otelgorm.NewPlugin()` instruments GORM only when that flag is true.
- OTLP uses an insecure gRPC connection and a batch span processor.
- Sampling is hard-coded to `AlwaysSample`.
- W3C trace context and baggage propagators are configured.
- `telemetry.StartSpan` exists but has no live call sites outside its own
  declaration.
- No Asynq tracing instrumentation is wired.
- Logs do not include OTEL `trace_id` or `span_id`.

### Realtime Stats

- `startMetricsBroadcaster` runs every two seconds.
- It calculates process-local RPS from an atomic request counter.
- It queries dashboard summary and system insights.
- It sends a `metrics_update` event to WebSocket channel `system:metrics`.
- The same loop prunes stale presence records.
- The loop uses `context.Background()` without deadlines or shutdown
  cancellation.
- `GetSystemInsights` returns hard-coded latency, error rate, uptime, and role.
- The broadcaster sends hard-coded CPU, memory, and active-thread values.

### Deployment

- Development Compose includes API, Jaeger, Prometheus, Grafana, MySQL,
  Redis, Mailpit, and object storage.
- Prometheus scrapes `api:8080` and relies on the default `/metrics` path.
- Prometheus has no Basic Auth configuration.
- Grafana provisions Prometheus and one dashboard.
- Production Compose has no in-repo Prometheus, Grafana, Jaeger, OTEL
  Collector, or log collector.
- Production health checks call `/health`, while the live route is
  `/api/v1/health`.
- Production Compose sets `APP_ENV`; runtime configuration reads
  `SERVER_APP_ENV`.

## Configuration Simulation

### Viper Behavior

The repository uses `github.com/spf13/viper v1.21.0` and does not depend on an
environment-tag parser such as `caarlos0/env`. Therefore `env:` and
`envDefault:` struct tags do not load configuration.

Viper 1.21 `Unmarshal` builds its key set from `AllKeys()`. `AllKeys()` includes
explicitly bound environment keys (`v.env`) and defaults, but not arbitrary
variables discovered by `AutomaticEnv`. The repository already documents this
behavior in `envOnlyKeys`.

Current consequences:

- Metrics work because defaults register their Viper keys and lines 402-405 of
  `internal/config/config.go` explicitly assign `v.Get*` values.
- Telemetry has neither Viper defaults nor explicit `v.Get*` assignments.
- `OTEL_ENABLED`, `OTEL_SERVICE_NAME`, and `OTEL_COLLECTOR_URL` are therefore
  not proven to populate `cfg.Telemetry` and are expected to be dropped.
- The correct Viper namespace for documented `OTEL_*` variables is `otel.*`,
  because the environment replacer maps `otel.enabled` to `OTEL_ENABLED`.
- Metrics credentials actually map from `metrics.username` and
  `metrics.password`, which means the effective variables are
  `METRICS_USERNAME` and `METRICS_PASSWORD`. The `METRICS_USER` and
  `METRICS_PASS` struct tags are not active.

### Logging Correlation Simulation

`TraceContextHook` reads `constants.UserIDKey`. Authentication writes user
identity through `authcontext.WithUserID`, whose private key is a different
typed context key. These values cannot match. Even after fixing the key,
entries that do not use `WithContext(ctx)` cannot be enriched.

OTEL's Gin middleware runs before Request ID middleware. The request context
available to `RequestLogger` can contain both a span and request ID, so a
Logrus hook can safely extract `trace_id`, `span_id`, and `request_id` from the
same context when tracing is enabled.

### Metrics Correctness Simulation

Removing placeholder fields is safer than substituting host-level collectors
inside the API process:

- Prometheus already exports Go runtime/process metrics.
- CPU and memory should be queried in Grafana or through a real metrics API,
  not hard-coded in an authenticated business endpoint.
- The WebSocket payload can retain values that are real today: process-local
  RPS, local active connections, and database-backed total users.
- Process-local RPS and active connections must be labeled as instance-local
  until a cluster aggregation design exists.

Changing or removing realtime fields is a cross-stack contract change and must
update `apps/client`, `apps/web`, and any shared API types in one slice.

## Confirmed Defects

### P0

1. OTEL environment configuration is not registered and assigned like the
   other Viper-backed configuration.
2. Production health checks use a route that does not exist.
3. Production uses `APP_ENV` while the application reads `SERVER_APP_ENV`.
4. User-visible system metrics contain hard-coded placeholder values.
5. Error-rate units are interpreted differently by frontend consumers.

### P1

1. Logs cannot be directly correlated with OTEL traces.
2. Automatic `user_id` enrichment reads the wrong context key.
3. OTEL is always sampled and always insecure.
4. Metrics are public by default, including in non-local environments unless
   deployment networking blocks access.
5. Prometheus scrape config has no support for backend Basic Auth.
6. Worker visibility is limited to logs and cleanup counters.
7. Metrics broadcaster has no managed shutdown or query deadlines.
8. Grafana lacks latency percentiles, error ratio, scrape health, and worker
   observability.

### P2

1. HTTP logs use raw paths and can create high-cardinality indexes.
2. Two recovery middleware implementations are installed.
3. The Next.js proxy does not return backend `X-Request-ID` to clients.
4. Frontend logging is unstructured and inconsistent.
5. Some worker and domain logs include direct personal or session identifiers.

## Existing Verification

The following command passed during the audit:

```bash
go test ./internal/config ./internal/middleware ./internal/router ./pkg/telemetry
```

`pkg/telemetry` currently has no tests. Existing middleware tests verify basic
request logging, status severity, latency fields, Prometheus middleware does
not panic, recovery behavior, trusted proxies, and metrics-auth validation.
They do not verify OTEL environment loading, trace/log correlation, metric
samples, route authorization for `/metrics`, or broadcaster lifecycle.

`docker compose -f docker-compose.dev.yml config` also rendered successfully,
with a local warning that `MYSQL_ROOT_PASSWORD` was unset.

## Decisions Required Before Execution

1. Placeholder metrics: recommended decision is to remove unavailable fields
   from API/WebSocket contracts rather than retain misleading values.
2. Production metrics security: recommended decision is fail-fast when metrics
   are enabled without authentication outside local/dev/test. An internal-only
   listener is a valid later alternative but is broader than the current
   router design.
3. OTEL transport: local development should remain insecure; non-local
   environments should require TLS unless explicitly overridden.

## Related Plans

- `llm/plans/improve/observability-p0.md`
- `llm/plans/improve/observability-p1.md`
- `llm/plans/improve/observability-p2.md`
- `llm/tasks/observability-progress.md`
