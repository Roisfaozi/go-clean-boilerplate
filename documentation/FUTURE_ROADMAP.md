# Future Roadmap & Improvements

This document outlines the strategic plan to elevate the project to a "World Class Enterprise Standard". It focuses on observability, scalability, and developer experience.

## 🟢 Phase 1: Observability & Monitoring (In Progress)
Essential for running the application in a production environment confidently.

- [x] **Structured Logging & Tracing**
    - [x] Add `RequestID` (Trace ID) to all log entries context.
    - [x] Implement Context-Aware Logging across UseCases and Repositories.
    - [ ] Implement OpenTelemetry (OTEL) for distributed tracing across layers.
- [x] **Health Checks Pro**
    - [x] Enhance `/health` to check DB ping and Redis connectivity status deeply.
- [ ] **Metrics Collection (Prometheus)**
    - [ ] Expose `/metrics` endpoint.
    - [ ] Track HTTP metrics: Request Rate (RPS), Latency (P95/P99), Error Rate.
    - [ ] Track Runtime metrics: Goroutines, Memory Usage, GC Duration.
    - [ ] Track Business metrics: Active Websocket connections, Registered Users count.

## 🟡 Phase 2: Scalability & Async Processing
Decouple heavy tasks from the main HTTP request flow.

- [ ] **Background Job System**
    - [ ] Integrate a Task Queue library (Recommended: `hibiken/asynq` using Redis).
    - [ ] Create a `Worker` server entry point (`cmd/worker/main.go`).
- [ ] **Async Audit Logging**
    - [ ] Move Audit Log creation from Synchronous UseCase calls to Background Jobs to improve API latency.
- [ ] **Email/Notification Service**
    - [ ] Create a dedicated module for sending emails (Welcome, Reset Password).
    - [ ] Process email sending asynchronously via the Task Queue.

## 🔵 Phase 3: Feature Expansion
Common enterprise features required by most applications. 
> 📄 **See [Module Improvement Specification](./MODULE_IMPROVEMENTS.md) for detailed feature breakdown.**

- [ ] **Auth Module Enhancements**
    - [ ] Forgot/Reset Password Flow (High).
    - [ ] Email Verification (Medium).
    - [ ] OAuth2 Social Login (Low).
- [ ] **User Module Enhancements**
    - [ ] User Status (Ban/Suspend) (High).
    - [ ] Profile Avatar Upload (Medium).
- [ ] **File Storage Module**
    - [ ] Create generic `StorageProvider` interface (Local/S3).
    - [ ] Add endpoint for uploading/serving user avatars.
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
