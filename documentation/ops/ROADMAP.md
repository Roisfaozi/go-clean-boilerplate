# Future Roadmap & Improvements

This document outlines the strategic plan to elevate the project to a "World Class Enterprise Standard". It focuses on observability, scalability, and developer experience.

## 🟢 Phase 1: Observability & Monitoring (Completed)

Essential for running the application in a production environment confidently.

- [x] **Structured Logging & Tracing**
  - [x] Add `RequestID` (Trace ID) to all log entries context.
  - [x] Implement Context-Aware Logging across UseCases and Repositories.
  - [x] Implement OpenTelemetry (OTEL) for distributed tracing across layers.
- [x] **Health Checks Pro**
  - [x] Enhance `/health` to check DB ping and Redis connectivity status deeply.
- [x] **Metrics Collection (Prometheus)**
  - [x] Expose `/metrics` endpoint.
  - [x] Track HTTP metrics: Request Rate (RPS), Latency (P95/P99), Error Rate.
  - [x] Track Runtime metrics: Goroutines, Memory Usage, GC Duration.
  - [x] Track Business metrics: Active Websocket connections, Registered Users count.

## 🟡 Phase 2: Scalability & Async Processing (Completed)

Decouple heavy tasks from the main HTTP request flow.

- [x] **Background Job System**
  - [x] Integrate a Task Queue library (Recommended: `hibiken/asynq` using Redis).
  - [x] Implement Scheduler for periodic maintenance tasks.
- [x] **Async Audit Logging**
  - [x] Support for audit logging within UseCases.
- [x] **Email/Notification Service**
  - [x] Process email sending asynchronously via the Task Queue (Simulation).

## 🔵 Phase 3: Feature Expansion (In Progress)

Common enterprise features required by most applications.

> 📄 **See [Module Improvement Specification](./MODULE_IMPROVEMENTS.md) for detailed feature breakdown.**

- [x] **Auth Module Enhancements**
  - [x] Forgot/Reset Password Flow (High).
  - [ ] Account Verification (Medium).
  - [ ] OAuth2 Social Login (Low).
- [x] **User Module Enhancements**
  - [x] User Status (Ban/Suspend) (High).
  - [x] Profile Avatar Upload (Medium).
- [x] **File Storage Module**
  - [x] Create generic `StorageProvider` interface (Local/S3).
  - [x] Add endpoint for uploading/serving user avatars.
- [x] **Multi-Tenancy (Organization Module)**
  - [x] Organization entity and repository with GORM scopes.
  - [x] Organization member management (owner, admin, member roles).
  - [x] TenantMiddleware for automatic organization context isolation.
  - [x] Organization CRUD API endpoints (create, read, update, delete).
  - [x] Complete Swagger documentation (39 total endpoints).
  - [ ] Tenant-scoped data isolation for existing modules (User, Role, Access, Audit).
  - [ ] Organization invitation system with email notifications.
  - [ ] Organization-level settings and configuration management.
- [ ] **Advanced Security (Remaining)**
  - [ ] **Rate Limiting Granularity**: Upgrade to Per-IP/Per-User limits.
  - [ ] **Circuit Breaker**: Implement `gobreaker` for external calls.
  - [ ] **MFA**: Add TOTP support.

## 🟣 Phase 4: Developer Experience (DX) & Tooling

Improve the speed and quality of development.

- [ ] **Pre-commit Hooks**
  - [ ] Configure `githooks` to run `make lint` & `make test-unit`.
- [ ] **Dependency Injection Refactor**
  - [ ] Refactor `internal/config/app.go` for better modularity.
- [ ] **API SDK Generator**
  - [ ] Auto-generate TypeScript/Axios client from Swagger.

---

## 📅 Completed Milestones (Latest)

- [x] **Observability**: Implemented Request ID Tracing (Middleware + Logrus Hook) and Deep Health Checks (MySQL/Redis).
- [x] **Core Upgrade**: Go 1.25.5 security patches.
- [x] **Architecture**: Clean Architecture with modular structure.
- [x] **Auth**: JWT with Redis-backed sessions.
- [x] **RBAC**: Casbin integration with Database persistence.
- [x] **Testing**: Full suite (Unit, Integration with Singleton Containers, E2E) with 100% Pass Rate.
- [x] **CI/CD**: GitHub Actions workflow.
- [x] **Security Hardening**:
  - [x] **Trusted Proxies**: Configured to prevent IP Spoofing (Fail-Fast mode).
  - [x] **Rate Limiting**: Atomic execution via Redis Lua script to prevent race conditions.
  - [x] **WebSocket**: Strict Origin validation to prevent CSWSH attacks.
  - [x] **Dynamic Query**: Blacklist sensitive fields (password, token) from sorting/filtering.
  - [x] **Input Validation**: Strict Regex for emails and password strength checks.
