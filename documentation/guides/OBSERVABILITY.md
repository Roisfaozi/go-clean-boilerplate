# Observability Guide (OpenTelemetry & Jaeger)

This project integrates **OpenTelemetry (OTEL)** to provide deep visibility into request lifecycles across different layers: HTTP (Gin), Database (GORM), and Background Workers (Asynq).

## Architecture

1.  **OTEL SDK**: Initialized at application startup in `internal/config/app.go`.
2.  **Instrumentation**:
    *   **Gin**: Middleware tracks incoming requests and generates `TraceID`.
    *   **GORM**: Plugin tracks SQL query execution time and errors.
    *   **Propagation**: The `TraceID` is automatically passed via `context.Context`.
3.  **Exporter**: Data is sent via gRPC to an OTLP-compatible collector (Jaeger).

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
-   **Identify Bottlenecks**: See exactly which SQL query or UseCase method is slowing down a request.
-   **Error Correlation**: View logs and traces together to understand why a request failed.
-   **Distributed Tracing**: If the project grows into microservices, the `TraceID` will follow the request across network boundaries.
