# System Architecture & Technical Guide

This document provides a deep dive into the architecture, technologies, and core patterns used in the Go Clean Boilerplate API.

---

## 1. Clean Architecture Overview

The project follows **Clean Architecture** principles to ensure independence from frameworks, databases, and external tools.

### Layered Structure
-   **Entities**: Core business objects (e.g., `User`, `Role`). Located in `internal/modules/*/entity/`.
-   **Use Cases**: Application-specific business logic. Coordinates data flow. Located in `internal/modules/*/usecase/`.
-   **Interface Adapters**:
    -   **Repositories**: Data access (GORM, Redis).
    -   **Controllers**: HTTP Handlers (Gin).
-   **Frameworks & Drivers**: The outermost layer (Gin, MySQL, Redis, Asynq).

---

## 2. Core Modules

### 🔐 Authentication & Authorization
-   **JWT**: Stateless authentication with stateful refresh tokens stored in Redis for instant revocation.
-   **Server-Side Auth Proxy**: All client-side requests are proxied through Next.js server routes.
    -   **Cookies**: Tokens are stored in `HttpOnly` secure cookies.
    -   **Proxy Injection**: The Next.js proxy layer automatically injects the `Authorization: Bearer` header before forwarding requests to the Go backend.
-   **Casbin (RBAC)**: Fine-grained access control.
    -   **Hierarchical Roles**: `role:superadmin` inherits `role:admin`.
    -   **Dynamic Policies**: Permission rules are stored in the database.
    -   **Access Rights**: Logical grouping of multiple physical API endpoints.

### 🌐 Frontend Orchestration (Next.js 16)
-   **`proxy.ts`**: The modern replacement for middleware, handling server-side route protection and internationalization at the edge.
-   **Server Actions**: Login and Logout flows are implemented as Server Actions to handle sensitive cookie operations securely.
-   **AuthProvider**: A client-side synchronization layer that hydrates the user's state and permissions from the backend on mount.

### 📊 Observability (OTEL)
-   **Tracing**: Distributed tracing via OpenTelemetry and Jaeger.
-   **Metrics**: Real-time performance monitoring via Prometheus and Grafana.
-   **Audit Logs**: Automatic tracking of sensitive operations (Create, Update, Delete).

### 🛠 Background Workers
-   **Engine**: Powered by `hibiken/asynq` (Redis-based).
-   **Scheduler**: Automated maintenance (Token pruning, Soft-delete cleanup).
-   **Async Tasks**: Email simulation and heavy processing.

### 📁 Storage Abstraction
-   **Strategy Pattern**: Switch between `local` disk and `s3`-compatible providers (MinIO, R2) via config.

---

## 3. Communication Patterns

-   **Synchronous**: Standard RESTful API via Gin.
-   **Asynchronous**: Event-driven background tasks via Asynq.
-   **Real-time**: 
    -   **WebSockets**: Bidirectional, distributed scaling via Redis Pub/Sub.
    -   **SSE**: Lightweight one-way server push.

---

## 4. Testing Standards

We employ a 3-layer testing strategy:
1.  **Unit Tests**: Isolated logic using `mockery`.
2.  **Integration Tests**: Real DB/Redis using **Singleton Testcontainers**.
3.  **E2E Tests**: Full HTTP flow validation.

*See [TESTING.md](./guides/TESTING.md) for detailed patterns.*
