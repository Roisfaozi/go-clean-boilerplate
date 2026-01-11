# Comprehensive Testing Strategy Guide

This document defines the official testing standards for this project, covering Unit, Integration, and E2E testing layers.

---

## 🏗️ 1. Unit Testing (Isolated Logic)
**Goal:** Verify business logic in isolation with zero external dependencies.

*   **Libraries**: `testify`, `mockery`.
*   **Pattern**: **Dependency Struct Pattern** for UseCases.

### Standard Setup
```go
type userTestDeps struct {
    Repo     *mocks.MockUserRepository
    TM       *mocking.MockWithTransactionManager
}

func setupUserTest() (*userTestDeps, usecase.UserUseCase) {
    deps := &userTestDeps{
        Repo: new(mocks.MockUserRepository),
        TM:   new(mocking.MockWithTransactionManager),
    }
    uc := usecase.NewUserUseCase(deps.TM, log, deps.Repo, ...)
    return deps, uc
}
```

---

## 🔗 2. Integration Testing (Real Infrastructure)
**Goal:** Verify interaction with real MySQL and Redis instances using **Singleton Containers**.

*   **Optimization**: We use the **Singleton Container Pattern** to start Docker only once per test suite.
*   **Cleanup**: Use `TRUNCATE` between tests instead of restarting containers.

### Execution
```bash
make test-integration
```

---

## 🌍 3. End-to-End (E2E) Testing (Full Flow)
**Goal:** Verify the complete HTTP request-response cycle from the client's perspective.

*   **Setup**: Uses `httptest.Server` connected to the singleton integration containers.
*   **Client**: Uses a custom `TestClient` wrapper for easy JSON assertions.

### Execution
```bash
make test-e2e
```

---

## 🛡️ 4. Security Testing
Every module must include:
- **SQL Injection Tests**: Ensuring inputs are parameterized.
- **RBAC Tests**: Verifying Casbin policies correctly block unauthorized roles.
- **Validation Tests**: Checking for invalid formats and required fields.

---

## ⚡ 5. Generating Mocks
When you modify an interface, you MUST regenerate mocks:
```bash
make mocks
```
