# Observability Guide (OpenTelemetry, Jaeger & Prometheus)

This project integrates **OpenTelemetry (OTEL)** and **Prometheus** to provide deep visibility into request lifecycles across different layers: HTTP (`net/http` via `otelhttp` and `HTTPPrometheusMiddleware`), Database (`sqlx` via `otelsql` and `NewDBStatsCollector`), Cache (`redis` via hook metrics), Uploads/Storage (`tus` metrics), Realtime (`NewRealtimeStatsCollector`), and Background Workers (`Asynq`, webhook/email/outbox metrics).

## Architecture

1.  **OTEL SDK & Metrics**: Initialized at application startup via `internal/config/fx.go` (`provideTracerShutdown`, `registerPrometheusCollectors`) and `internal/config/app.go`.
2.  **Instrumentation**:
    - **HTTP (`otelhttp` + `HTTPPrometheusMiddleware`)**: Wraps `net/http` handlers to track incoming requests, generate `TraceID`, and record Prometheus HTTP metrics, including rate-limit rejection counters (`app_rate_limit_rejections_total`).
    - **Database (`otelsql` + `NewDBStatsCollector`)**: Instruments `sqlx` (`internal/config/sqlx.go`) SQL query execution spans/errors and exports DB connection pool stats.
    - **Cache & Redis**: Records operation counts and latency histograms via Redis client hooks (`app_redis_operations_total`, `app_redis_operation_duration_seconds`).
    - **Uploads & Storage (`tus`)**: Records upload file sizes, duration, and completion statuses (`app_tus_upload_size_bytes`, `app_tus_upload_duration_seconds`, `app_storage_uploads_total`).
    - **Realtime & Workers**: `NewRealtimeStatsCollector` exports WebSocket/SSE stats, alongside webhook, email, and outbox metrics (`app_audit_outbox_oldest_age_seconds`, `app_audit_outbox_pending`).
    - **Propagation**: The `TraceID` is automatically passed via `context.Context`.
3.  **Exporter**: Traces are sent via gRPC to an OTLP-compatible collector (Jaeger), and Prometheus metrics are exposed at `/metrics`.

## Prometheus Metrics

The application exposes Prometheus metrics at `/metrics` (with endpoint security in non-dev environments). Key metrics include:

| Metric Name | Type | Description | Labels |
|---|---|---|---|
| `app_redis_operations_total` | Counter | Total Redis operations by bounded command and status | `cmd`, `status` |
| `app_redis_operation_duration_seconds` | Histogram | Duration of Redis operations in seconds | `cmd`, `status` |
| `app_audit_outbox_oldest_age_seconds` | Gauge | Age in seconds of the oldest pending audit outbox entry in batch | None |
| `app_audit_outbox_pending` | Gauge | Number of pending audit outbox entries in batch | None |
| `app_rate_limit_rejections_total` | Counter | Total HTTP requests rejected by rate limiters | `limiter` |
| `app_tus_upload_size_bytes` | Histogram | Size of completed TUS uploads in bytes | `driver` |
| `app_tus_upload_duration_seconds` | Histogram | Duration of TUS upload completion hook handling in seconds | `driver`, `status` |
| `app_storage_uploads_total` | Counter | Total number of file uploads | `driver`, `status` |
| `app_user_logins_total` | Counter | Total number of user login attempts | `status` |
| `app_user_registrations_total` | Counter | Total number of new user registrations | None |
| `app_websocket_active_connections` | Gauge | Number of currently active WebSocket connections | None |
| `app_ws_send_drops_total` | Counter | Outbound WebSocket messages dropped due to full buffer | None |
| `app_worker_tasks_total` | Counter | Total background worker tasks processed | `task_type`, `status` |
| `app_worker_task_duration_seconds` | Histogram | Duration of background worker task processing | `task_type`, `status` |
| `app_webhook_deliveries_total` | Counter | Webhook delivery attempts by bounded outcome | `outcome` |
| `app_webhook_delivery_duration_seconds` | Histogram | Duration of webhook delivery attempts in seconds | `outcome` |
| `app_email_deliveries_total` | Counter | Email delivery attempts by bounded outcome | `outcome` |
| `app_email_delivery_duration_seconds` | Histogram | Duration of email delivery attempts in seconds | `outcome` |

## Setup & Visualization

### 1. Enable in .env

```env
OTEL_ENABLED=true
OTEL_SERVICE_NAME=go-clean-api
OTEL_COLLECTOR_URL=localhost:4317
```

### 2. Start Jaeger

The development environment includes a Jaeger container in `docker-compose.dev.yml`.

```bash
make docker-dev
```

### 3. Access Dashboard

Open **[http://localhost:16686](http://localhost:16686)** in your browser.

- Select your service name (default: `go-clean-api`).
- Click **Find Traces**.

## Manual Instrumentation

To add custom spans within your business logic (UseCase):

```go
import "github.com/Roisfaozi/go-clean-boilerplate/pkg/telemetry"

func (u *myUseCase) ComplexOperation(ctx context.Context) error {
    ctx, span := telemetry.StartSpan(ctx, "MyUseCase.ComplexOperation")
    defer span()

    // ... your logic ...
    return nil
}
```

## Benefits

- **Identify Bottlenecks**: See exactly which SQL query or UseCase method is slowing down a request.
- **Error Correlation**: View logs and traces together to understand why a request failed.
- **Distributed Tracing**: If the project grows into microservices, the `TraceID` will follow the request across network boundaries.
