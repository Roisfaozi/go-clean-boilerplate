# Observability P1 Plan - Production Readiness

## Objective

Connect logs, traces, metrics, workers, and lifecycle management into a usable
production observability path after P0 establishes truthful behavior.

## Prerequisites

- All P0 exit criteria are complete.
- OTEL environment loading is covered by tests.
- Realtime stats contain no placeholder fields.
- Production metrics exposure policy has been approved.

## P1-1 - Correlate Logs With Requests, Users, and Traces

### Files

- `internal/config/logrus.go`
- new `internal/config/logrus_test.go`
- `pkg/authcontext/context.go`
- context-aware log call sites on critical auth/tenant/worker paths

### Planned Change

1. In `TraceContextHook`, extract valid OTEL span context:
   - `trace_id`
   - `span_id`
   - optional `trace_sampled`
2. Read user ID through `authcontext.UserIDFromContext`, not
   `constants.UserIDKey`.
3. Keep request ID extraction unchanged unless a shared accessor is introduced.
4. Convert only critical request-owned logs that currently discard context to
   `WithContext(ctx)`; broad module cleanup belongs to P2.
5. Never put session tokens, API keys, cookies, reset tokens, or passwords in
   enrichment fields.

### Tests

- Entry with request context contains `request_id`.
- Auth context contains `user_id`.
- Valid sampled span contains `trace_id`, `span_id`, and sampled state.
- Background context emits none of those fields and does not panic.
- Invalid/non-recording span does not emit zero IDs.

### Acceptance Criteria

- A failed authenticated request can be followed from request log to domain log
  and trace by stable fields.
- Enrichment is deterministic and does not mutate unrelated entries.

## P1-2 - Make OTEL Transport and Sampling Environment-Aware

### Files

- `internal/config/config.go`
- `internal/config/config_test.go`
- `pkg/telemetry/tracer.go`
- new `pkg/telemetry/tracer_test.go`
- `internal/config/app.go`
- `.env.example`
- `documentation/guides/OBSERVABILITY.md`

### Proposed Config

```text
OTEL_ENABLED=false
OTEL_SERVICE_NAME=go-clean-api
OTEL_COLLECTOR_URL=localhost:4317
OTEL_INSECURE=true
OTEL_SAMPLE_RATIO=1.0
OTEL_EXPORT_TIMEOUT=10s
```

Future authentication headers or certificates should be added only when the
chosen collector contract is known.

### Planned Change

- Pass a typed tracer config instead of three loose arguments.
- Use parent-based trace ID ratio sampling.
- Reject sample ratios outside `[0,1]`.
- Permit insecure transport for local development.
- Require explicit insecure opt-in outside local/dev/test, or require TLS.
- Bound exporter initialization and shutdown with configured timeouts.
- Add service environment and version resource attributes when build metadata
  is available.

### Simulation

- Local: insecure Jaeger at `jaeger:4317`, ratio `1.0`.
- Production: TLS collector endpoint, conservative ratio.
- Disabled: no tracer provider or Gin/GORM instrumentation.
- Collector unavailable: startup policy remains non-fatal unless strict mode is
  explicitly approved; failures must be logged with config-safe fields.

## P1-3 - Secure Metrics Exposure

### Recommended Policy

- Local/dev/test: metrics may be unauthenticated on the local network.
- Staging/production: if metrics are enabled, authentication credentials are
  required and startup fails if they are absent.

### Files

- `internal/config/config.go`
- `internal/config/config_test.go`
- `internal/router/router.go`
- `internal/router/router_test.go`
- `deploy/prometheus/prometheus.yml`
- `.env.example`
- production deployment docs

### Planned Change

1. Move strict-environment classification to a shared config helper if needed.
2. Validate production metrics configuration during `NewConfig`.
3. Add router tests for disabled, unauthenticated local, unauthorized Basic
   Auth, and authorized Basic Auth cases.
4. Configure Prometheus `metrics_path`, `scrape_timeout`, and `basic_auth`.
5. Do not put literal production credentials in version control. Use mounted
   secrets or environment substitution supported by the deployment system.

### Alternative

A separate internal metrics listener is architecturally stronger but requires
additional server lifecycle and deployment wiring. Track it as a later design
if Basic Auth plus network isolation is insufficient.

## P1-4 - Add Worker Metrics

### Files

- `pkg/telemetry/metrics.go`
- `internal/worker/processor.go`
- worker tests
- Grafana dashboard

### Proposed Metrics

```text
app_worker_tasks_total{task_type,status}
app_worker_task_duration_seconds{task_type,status}
app_worker_task_failures_total{task_type}
```

Avoid task IDs, user IDs, queue payloads, URLs, and error messages as labels.
Queue depth should come from an Asynq inspector/exporter rather than a label on
task execution metrics.

### Planned Change

- Wrap registered handlers with one instrumentation function.
- Record started/success/failed and duration.
- Preserve handler return errors and Asynq retry behavior.
- Keep existing cleanup metrics during migration; remove duplicates only after
  dashboards and tests move to the generic worker metrics.
- Enqueue metrics are optional in the first slice because changing the
  distributor interface has a larger mock blast radius.

### Acceptance Criteria

- Every registered worker handler emits consistent outcomes.
- A retry is not reported as success.
- Metrics add bounded labels only.
- Existing handler semantics and tests remain unchanged.

## P1-5 - Manage Realtime Metrics Lifecycle

### Files

- `internal/config/app_helpers.go`
- `internal/config/app.go`
- tests around broadcaster lifecycle

### Planned Change

1. Give the broadcaster a cancellable context owned by `Application`.
2. Stop its ticker on cancellation.
3. Add per-query timeouts shorter than the broadcast interval.
4. Log query and serialization failures with structured fields.
5. Separate presence pruning from the two-second metrics loop.
6. Run presence pruning at a configured interval matching its actual intent.
7. Stop both loops before database and Redis shutdown.
8. Add a wait group or explicit done channel so shutdown can wait within its
   deadline.

### Acceptance Criteria

- No broadcaster query starts after cancellation.
- Shutdown waits for or times out the loop deterministically.
- A stats failure does not silently stop all future broadcasts.
- Presence pruning cadence is independent of metrics cadence.

## P1-6 - Improve Grafana Operational Views

### Files

- `deploy/grafana/dashboards/main_dashboard.json`
- optional new alert rules under `deploy/prometheus/`
- provisioning files if alert data sources are added

### Panels

- Request rate using `rate`, split by instance and route as needed.
- p50/p95/p99 latency using histogram buckets.
- 4xx and 5xx rate and 5xx ratio.
- scrape target health (`up`).
- active WebSocket connections.
- login success/failure rate.
- storage and cleanup/worker failures.
- Go heap and goroutines with instance labels.

### Alerts

Start with evidence-based operational alerts, not arbitrary business SLOs:

- target down for five minutes;
- elevated 5xx ratio over a sustained window;
- p95 latency above an agreed threshold;
- repeated worker failures;
- Prometheus scrape failures.

Thresholds require owner approval and observed baseline data.

## P1 Verification Matrix

```bash
go test ./internal/config ./internal/router ./internal/worker/... ./pkg/telemetry -count=1
go test ./internal/... ./pkg/...
docker compose -f docker-compose.dev.yml config
docker compose -f docker-compose.prod.yml config
promtool check config deploy/prometheus/prometheus.yml
```

When Docker is available, start the development stack and verify:

1. `/metrics` rejects missing credentials under strict configuration.
2. Prometheus target is up.
3. Jaeger receives Gin and GORM spans.
4. JSON logs contain matching request, trace, span, and user IDs.
5. A deliberately failed worker task increments failure metrics.
6. Grafana panels return non-empty data.

## P1 Exit Gate

- Logs and traces are correlated by tests and a manual request.
- OTEL sampling and transport are configurable and validated.
- Metrics exposure policy is enforced.
- Prometheus can scrape the protected endpoint.
- Worker outcomes are measurable.
- Realtime loops shut down deterministically.
- Dashboard panels use real metric queries.
