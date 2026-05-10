# System Architecture & Technical Guide

This document provides a deep dive into the architecture, technologies, and core patterns used in the Go Clean Boilerplate API.

---

## 1. Clean Architecture Overview

The project follows **Clean Architecture** principles to ensure independence from frameworks, databases, and external tools.

### Layered Structure

- **Entities**: Core business objects (e.g., `User`, `Role`). Located in `internal/modules/*/entity/`.
- **Use Cases**: Application-specific business logic. Coordinates data flow. Located in `internal/modules/*/usecase/`.
- **Interface Adapters**:
  - **Repositories**: Data access (GORM, Redis).
  - **Controllers**: HTTP Handlers (Gin).
- **Frameworks & Drivers**: The outermost layer (Gin, MySQL, Redis, Asynq).

---

## 2. Core Modules

### 🔐 Authentication & Authorization

- **JWT**: Stateless authentication with stateful refresh tokens stored in Redis for instant revocation.
- **Casbin (RBAC)**: Fine-grained access control.
  - **Hierarchical Roles**: `role:superadmin` inherits `role:admin`.
  - **Dynamic Policies**: Permission rules are stored in the database.
  - **Access Rights**: Logical grouping of multiple physical API endpoints.

### 📊 Observability (OTEL)

- **Tracing**: Distributed tracing via OpenTelemetry and Jaeger.
- **Metrics**: Real-time performance monitoring via Prometheus and Grafana.
- **Audit Logs**: Automatic tracking of sensitive operations (Create, Update, Delete).

### 🛠 Background Workers

- **Engine**: Powered by `hibiken/asynq` (Redis-based).
- **Scheduler**: Automated maintenance (Token pruning, Soft-delete cleanup).
- **Async Tasks**: Email simulation and heavy processing.

### 🏢 Multi-Tenancy (Organizations)

- **Isolation**: Data is isolated at the organization level using `tenant_middleware`.
- **Membership**: Users can belong to multiple organizations with different individual roles.

### 📋 Resource Management

- **Projects**: Grouping of resources within an organization. Access is governed by organization membership.
- **Stats & Insights**: Real-time aggregation of activity and resource usage.

### 📁 Storage Abstraction

- **Strategy Pattern**: Switch between `local` disk and `s3`-compatible providers (MinIO, R2) via config.
- **TUS Support**: Robust, resumable file uploads integrated into the storage layer.

---

## 3. Frontend Architecture (Next.js)

The project includes a modern frontend built with **Next.js 14**, providing a seamless admin dashboard experience.

- **Tech Stack**: Next.js (App Router), TypeScript, Tailwind CSS, Shadcn UI.
- **API Integration**: Typed API client with automatic organization context injection via headers (`X-Organization-ID`).
- **Authentication**: Session management synchronized with Go backend via HttpOnly cookies and refresh token rotation.

---

## 3. Communication Patterns

- **Synchronous**: Standard RESTful API via Gin.
- **Asynchronous**: Event-driven background tasks via Asynq.
- **Real-time**:
  - **WebSockets**: Bidirectional, distributed scaling via Redis Pub/Sub.
  - **SSE**: Lightweight one-way server push.

---

## 4. Testing Standards

We employ a 3-layer testing strategy:

1.  **Unit Tests**: Isolated logic using `mockery`.
2.  **Integration Tests**: Real DB/Redis using **Singleton Testcontainers**.
3.  **E2E Tests**: Full HTTP flow validation.

_See [TESTING.md](./guides/TESTING.md) for detailed patterns._

---

## 5. Design Decisions & Standards

### 🔄 Avoiding Circular Dependencies

To prevent dependency cycles (e.g., `config -> middleware -> usecase -> config`), UseCase constructors **MUST NOT** accept the full `internal/config.AppConfig` struct. Instead, pass the specific required values (strings, ints, etc.) as raw arguments.

### 📁 Storage Context Propagation

All methods in `pkg/storage.Provider` MUST accept `context.Context`. This ensures that file operations respect request deadlines, cancellations, and allow for proper trace propagation.

### 🛡 Casbin Transactional Integrity

Casbin operations must be wrapped in the `TransactionalEnforcer` when part of a GORM transaction. The enforcer MUST propagate the transaction handle via `.WithContext(ctx)` to ensure that policy changes are committed or rolled back atomically with other database operations.
