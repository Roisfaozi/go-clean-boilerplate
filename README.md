# Go Clean Boilerplate - Enterprise Modular REST API

![Go Version](https://img.shields.io/badge/Go-1.25.5%2B-blue)
![License](https://img.shields.io/badge/License-Apache%202.0-green)
![Architecture](https://img.shields.io/badge/Architecture-Clean%20%26%20Modular-orange)
![Testing](https://img.shields.io/badge/Testing-Unit%2C%20Integration%2C%20E2E-success)
![Realtime](https://img.shields.io/badge/Realtime-Distributed%20WS%20%26%20SSE-ff69b4)

Enterprise-ready Go boilerplate implementing Clean Architecture, RBAC with Casbin, Modular Audit Logging, and Distributed WebSocket scaling.

---

## 🚀 Core Features

-   **Clean Architecture**: Strict separation of concerns (Delivery, UseCase, Repository, Entity).
-   **Advanced RBAC with Casbin**: Fine-grained access control using GORM adapter. Policies are stored in the database for dynamic updates.
-   **Distributed WebSockets**: Scalable WebSocket management using **Redis Pub/Sub** backplane, allowing multi-node synchronization.
-   **Modular Audit Logging**: Synchronous activity tracking (LOGIN, REGISTER, UPDATE, DELETE) integrated directly into business UseCases.
-   **Multi-Provider File Storage**: Pluggable storage abstraction supporting **Local Disk**, **AWS S3**, **MinIO**, and **Cloudflare R2**.
-   **Automated Cleanup Jobs**: Integrated background worker scheduler for database maintenance (token pruning, soft-delete cleanup, log rotation).
-   **Distributed Tracing (OTEL)**: Full visibility with **OpenTelemetry** integration, tracking request flow across HTTP, Database, and Workers.
-   **Dynamic Search & Filtering**: Secure, reusable query builder supporting complex clauses, range filters, and dynamic sorting.
-   **Secure Authentication**: JWT-based auth with stateful session management in Redis for instant token revocation.
-   **Real-time SSE**: Server-Sent Events manager for live one-way data streaming.
-   **Hardened Security**: 
    -   UseCase-level validation (Regex email, password strength).
    -   Automatic HTTP security headers.
    -   Go 1.25.5 for critical security fixes.
-   **Comprehensive Testing**:
    -   **Unit Tests**: Fast, mock-based verification of logic.
    -   **Integration Tests**: Lightweight testing using **Singleton Testcontainers** pattern.
    -   **E2E Tests**: Full HTTP lifecycle validation.

---

## 🎛️ Toggleable Features & Configuration

This project is designed with high flexibility. Many core features can be enabled/disabled via environment variables (`.env`).

### Core Features

| Feature | Env Variable | Default | Description |
| :--- | :--- | :---: | :--- |
| **RBAC Authorization** | `CASBIN_ENABLED` | `false` | Enables Casbin authorization checks. If `false`, authorization is bypassed. |
| **Casbin Sync** | `CASBIN_WATCHER_ENABLED` | `false` | Enables policy sync across instances via Redis. Required for multi-replica setups. |
| **Rate Limiter** | `RATE_LIMIT_ENABLED` | `true` | Limits requests per second (RPS) per IP to prevent DoS/Brute Force. |
| **Distributed WS** | `WEBSOCKET_DISTRIBUTED_ENABLED` | `false` | Enables WebSocket message sync via Redis Pub/Sub. Required for horizontal scaling. |

### Security & Network

| Configuration | Env Variable | Default | Description |
| :--- | :--- | :---: | :--- |
| **Trusted Proxies** | `SERVER_TRUSTED_PROXIES` | *Empty* | Comma-separated list of trusted Load Balancer IPs/CIDRs. |
| **CORS Origins** | `CORS_ALLOWED_ORIGINS` | `*` | Allowed domains for CORS. |
| **JWT Secrets** | `JWT_ACCESS_SECRET`<br>`JWT_REFRESH_SECRET` | - | **Critical**: Must be random strings (min 32 chars). |

### Telemetry & Observability

| Configuration | Env Variable | Default | Description |
| :--- | :--- | :---: | :--- |
| **OTEL Tracing** | `OTEL_ENABLED` | `false` | Enables OpenTelemetry tracing. |
| **OTEL Service** | `OTEL_SERVICE_NAME` | `go-clean-api` | Service name shown in Jaeger/Tempo. |
| **Collector URL** | `OTEL_COLLECTOR_URL` | `localhost:4317` | OTLP gRPC collector endpoint (e.g. Jaeger). |

### Storage

| Configuration | Env Variable | Default | Description |
| :--- | :--- | :---: | :--- |
| **Driver** | `STORAGE_DRIVER` | `local` | Storage strategy: `local` or `s3`. |
| **Root Path** | `STORAGE_LOCAL_ROOT_PATH` | `./uploads` | Local directory for file storage. |
| **S3 Endpoint** | `STORAGE_S3_ENDPOINT` | - | Custom S3 endpoint (required for MinIO/R2). |

### Performance

| Configuration | Env Variable | Default | Description |
| :--- | :--- | :---: | :--- |
| **Rate Limit Store** | `RATE_LIMIT_STORE` | `memory` | Counter storage: `memory` (single instance) or `redis` (distributed). |
| **WS Ping Period** | `WEBSOCKET_PING_PERIOD` | *Auto* | Keep-alive ping interval (default: 90% of Pong Wait). |

---

## 📦 Deployment Scenarios

Choose the configuration that matches your infrastructure scale.

### 1. Single Instance (Monolith)
Suitable for development, small VPS, or simple deployments.

```env
# No need for distributed sync
RATE_LIMIT_STORE=memory
WEBSOCKET_DISTRIBUTED_ENABLED=false
CASBIN_WATCHER_ENABLED=false
```

### 2. Distributed Cluster (Kubernetes/Load Balanced)
Suitable for high-availability setups with multiple API replicas. Requires a shared Redis instance.

```env
# Sync state via Redis
RATE_LIMIT_STORE=redis
WEBSOCKET_DISTRIBUTED_ENABLED=true
CASBIN_WATCHER_ENABLED=true

# Security behind Load Balancer
SERVER_TRUSTED_PROXIES=10.0.0.0/8,172.16.0.0/12  # IPs of your LB/Ingress
```

---

## 🛠️ Technology Stack

| Category | Technology | Description |
| :--- | :--- | :--- |
| **Language** | [Go 1.25.5+](https://go.dev/) | Core programming language |
| **Framework** | [Gin Gonic](https://github.com/gin-gonic/gin) | High-performance HTTP framework |
| **Database** | [MySQL 8.0](https://www.mysql.com/) | Primary relational database |
| **Cache/Session** | [Redis 7](https://redis.io/) | Session storage & WS Pub/Sub backplane |
| **Authorization** | [Casbin](https://casbin.org/) | RBAC model & Policy enforcement |
| **Migrations** | [golang-migrate](https://github.com/golang-migrate/migrate) | Database schema management |
| **Testing** | [Testcontainers](https://testcontainers.com/) | Real instances for integration tests |

---

## 🏁 Getting Started

### Prerequisites
1.  **Go**: Version 1.25.5 or higher.
2.  **Docker & Docker Compose**: For running MySQL and Redis services easily.
3.  **Make**: For running automation commands defined in `Makefile`.
4.  **Air** (Optional): For live reloading during development.
    ```bash
    go install github.com/air-verse/air@latest
    ```
5.  **Swag CLI** (Optional): For regenerating API docs.
    ```bash
    go install github.com/swaggo/swag/cmd/swag@latest
    ```
6.  **Golang Migrate** (Optional): If you want to run migrations manually without the Makefile helper.


### Installation
1.  **Clone & Configure**:
    ```bash
    git clone https://github.com/Roisfaozi/go-clean-boilerplate.git
    cd go-clean-boilerplate
    cp .env.example .env
    ```
2.  **Start Infrastructure**:
    ```bash
    docker-compose up -d
    ```
3.  **Run Migrations & Seeding**:
    ```bash
    make migrate-up
    ```
4.  **Run Application**:
    ```bash
    make run
    ```

---

## 🧪 Testing Strategy

We use a layered testing strategy optimized for both speed and reliability.

| Command | Type | Description |
| :--- | :--- | :--- |
| `make test-unit` | **Unit** | Runs mock-based tests for internal/pkg logic. |
| `make test-integration` | **Integration** | Uses **Singleton Containers** for DB/Redis logic. |
| `make test-e2e` | **E2E** | Validates full HTTP request/response flows. |
| `make test-all` | **Full Suite** | Executes all test layers sequentially. |
| `make test-coverage` | **Coverage** | Generates an interactive HTML coverage report. |

> **Note**: Integration and E2E tests require Docker. We use a **Singleton Container Pattern** to reuse a single database/redis instance across the entire suite, drastically reducing execution time and resource usage.

---

## 📂 Project Structure

The project follows a standard Go project layout suitable for scalable microservices or monolithic APIs.

```
.
├── .air.toml           # Configuration for Air (live reloading)
├── Makefile            # Automation commands (build, test, migrate, run, mocks)
├── README.md           # Main project documentation
├── docker-compose.yml  # Docker services definition (MySQL, Redis)
├── go.mod              # Go dependency definitions
│
├── cmd/
│   └── api/            # Application entry point (main.go)
│
├── db/
│   ├── migrations/     # Database schema migration files (.sql)
│   └── seeds/          # Initial data seeding scripts (e.g. bootstrapping)
│
├── docs/               # Auto-generated Swagger/OpenAPI documentation files
│
├── documentation/      # Project guides and additional documentation
│   ├── USAGE.md                        # Detailed guide on how to use the API workflow
│   ├── DYNAMIC_SEARCH_EXAMPLES.md      # Curl examples for dynamic search
│   └── GET_VS_DYNAMIC_SEARCH.md        # Explains GET vs POST search approaches
│   └── SSE_USAGE.md                    # Guide for Server-Sent Events (SSE)
│
├── postman/            # Postman collections for testing
│   ├── Casbin Project API.postman_collection.json         # Main collection
│   ├── Casbin Project API - Dynamic Search.postman_collection.json # Dynamic search tests
│   ├── Casbin Project API - Realtime.postman_collection.json # Realtime features (WS, SSE)
│   └── ...
│
└───internal/           # Private application code (not importable by other apps)
    ├── config/         # Configuration loading & app initialization wiring
    ├── middleware/     # HTTP Middlewares (Auth, Casbin Enforcer, CORS, OTEL)
    ├── router/         # Gin router setup and route registration
    ├── worker/         # Background tasks, handlers & scheduler
    │
    └── modules/        # Domain-specific modules following Clean Architecture
        ├── auth/       # Authentication logic & JWT handling
        ├── user/       # User management (CRUD) & Avatar Upload
        ├── role/       # Role management
        ├── permission/ # Permission/Policy management (Casbin)
        └── access/     # Access Right & Endpoint management
```
## 📜 Documentation Links

- [System Architecture](./documentation/ARCHITECTURE.md)
- [API Usage Guide](./documentation/guides/API_USAGE.md)
- [API Access & RBAC](./documentation/API_ACCESS_WORKFLOW.md)
- [Testing Strategy](./documentation/guides/TESTING.md)
- [Real-time (WS & SSE)](./documentation/guides/REALTIME.md)
- [Dynamic Search](./documentation/guides/SEARCH.md)
- [Multi-Provider Storage](./documentation/guides/STORAGE.md)
- [Observability (Tracing/Metrics)](./documentation/guides/OBSERVABILITY.md)
- [Maintenance & Scheduler](./documentation/guides/MAINTENANCE.md)
- [Server Setup (Ops)](./documentation/ops/SERVER_SETUP_GUIDE.md)
- [Future Roadmap](./documentation/ops/ROADMAP.md)

---

## 📄 License

This project is licensed under the Apache 2.0 License - see the [LICENSE](LICENSE) file for details.