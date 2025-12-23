# Testing Strategy

Our testing philosophy focuses on multi-layer verification to ensure high confidence without sacrificing development speed.

## 🧱 The Testing Pyramid

### 1. Unit Tests (Base)
- **Scope**: Individual functions, struct methods, and business logic.
- **Mocking**: External dependencies (DB, Redis, WebSockets) are 100% mocked using `Mockery`.
- **Location**: Neighboring files (e.g., `user_usecase_test.go`).
- **Goal**: Rapid feedback during development.

### 2. Integration Tests (Middle)
- **Scope**: Repository logic, UseCase integration with real database constraints, and Casbin policy evaluation.
- **Infrastucture**: Uses **Singleton Docker Containers** via `testcontainers-go`.
- **Goal**: Ensure data integrity and correct SQL execution.

### 3. End-to-End (E2E) Tests (Top)
- **Scope**: Full HTTP request/response flow including Middleware, Routing, and Auth.
- **Mechanism**: Runs a real HTTP server using `httptest.Server` and a custom `TestClient`.
- **Goal**: Validate that the system works as a whole for the final user.

## 🏃 Execution Commands

| Target | Command | Notes |
| :--- | :--- | :--- |
| **All** | `make test-all` | Standard CI command. |
| **Business Logic** | `make test-unit` | No Docker required. |
| **DB / Casbin** | `make test-integration` | Requires Docker. |
| **API Endpoints** | `make test-e2e` | Real server environment. |

## 🛡 Security Testing
Security test cases are integrated into the Integration and E2E suites:
- **SQL Injection**: Payloads are tested against Auth and User inputs.
- **XSS**: Input sanitization/validation is verified.
- **Path Traversal**: File path inputs are hardened.
- **Rate Limiting**: Verified via E2E brute-force scenarios.