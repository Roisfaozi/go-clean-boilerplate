# Observability Guide (OpenTelemetry, Jaeger & Prometheus)

This project integrates **OpenTelemetry (OTEL)** and **Prometheus** to provide deep visibility into request lifecycles across different layers: HTTP (`net/http` via `otelhttp` and `HTTPPrometheusMiddleware`), Database (`sqlx` via `otelsql` and `NewDBStatsCollector`), Realtime (`NewRealtimeStatsCollector`), and Background Workers (`Asynq`, webhook/email/outbox metrics).

## Architecture

1.  **OTEL SDK & Metrics**: Initialized at application startup via `internal/config/fx.go` (`provideTracerShutdown`, `registerPrometheusCollectors`) and `internal/config/app.go`.
2.  **Instrumentation**:
    - **HTTP (`otelhttp` + `HTTPPrometheusMiddleware`)**: Wraps `net/http` handlers to track incoming requests, generate `TraceID`, and record Prometheus HTTP metrics.
    - **Database (`otelsql` + `NewDBStatsCollector`)**: Instruments `sqlx` (`internal/config/sqlx.go`) SQL query execution spans/errors and exports DB connection pool stats.
    - **Realtime & Workers**: `NewRealtimeStatsCollector` exports WebSocket/SSE stats, alongside webhook, email, and outbox metrics.
    - **Propagation**: The `TraceID` is automatically passed via `context.Context`.
3.  **Exporter**: Traces are sent via gRPC to an OTLP-compatible collector (Jaeger), and Prometheus metrics are exposed at `/metrics`.

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
