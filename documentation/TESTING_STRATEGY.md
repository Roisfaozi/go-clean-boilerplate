# Comprehensive Testing Strategy Guide

This document defines the **official testing standards** for the Go Clean Boilerplate project. It reflects the consolidated and refactored testing architecture currently in place.

---

## 🏗️ 1. Unit Testing (Isolated Logic)
**Goal:** Verify business logic in isolation with **zero** external dependencies.

*   **Scope:** `internal/modules/<module>/test/`
*   **Naming Convention:** `usecase_test.go`, `controller_test.go`
*   **Key Pattern:** **Dependency Struct Pattern** for UseCases.

### Standard: UseCase Test Structure
Every UseCase test file MUST define a dependency struct and a setup helper to keep tests clean.

```go
// internal/modules/user/test/use_case_test.go

// 1. Define Dependencies Struct
type userTestDeps struct {
    Repo     *mocks.MockUserRepository
    Enforcer *permMocks.IEnforcer
    AuditUC  *auditMocks.MockAuditUseCase
    // ... other mocks
}

// 2. Define Setup Helper
func setupUserTest() (*userTestDeps, usecase.UserUseCase) {
    deps := &userTestDeps{
        Repo:     new(mocks.MockUserRepository),
        Enforcer: new(permMocks.IEnforcer),
        AuditUC:  new(auditMocks.MockAuditUseCase),
    }
    // Initialize UseCase with Mocks
    uc := usecase.NewUserUseCase(deps.TM, log, deps.Repo, deps.Enforcer, deps.AuditUC)
    return deps, uc
}

// 3. Write Clean Tests
func TestUserUseCase_Create_Success(t *testing.T) {
    // Clean Setup
    deps, uc := setupUserTest()

    // Clear Expectations
    deps.Repo.On("FindByUsername", mock.Anything, "newuser").Return(nil, gorm.ErrRecordNotFound)
    deps.Repo.On("Create", mock.Anything, mock.Anything).Return(nil)

    // Execute
    resp, err := uc.Create(context.Background(), &model.RegisterUserRequest{...})

    // Assert
    assert.NoError(t, err)
    deps.Repo.AssertExpectations(t)
}
```

### Standard: Controller Test Structure
Use `setup<Module>TestRouter` and `newTest<Module>Controller` helpers.

```go
func TestUserHandler_Register_Success(t *testing.T) {
    mockUseCase := new(mocks.MockUserUseCase)
    handler := newTestUserController(mockUseCase) // Factory helper
    router := setupUserTestRouter()               // Router helper (gin.New)
    router.POST("/users", handler.Register)

    // ... expectations and execution ...
}
```

---

## 🔗 2. Integration Testing (Real Infrastructure)
**Goal:** Verify interaction with real MySQL and Redis instances using **Singleton Containers**.

*   **Scope:** `tests/integration/modules/`
*   **Naming Convention:** `<module>_integration_test.go` (Single file per module, NO `_comprehensive` suffixes).
*   **Key Pattern:** **Module Setup Helper** injection.

### Standard: Integration Test Structure

Every module integration test MUST use a setup helper that initializes the real Repository and UseCase using the shared `TestEnvironment`.

```go
// tests/integration/modules/auth_integration_test.go

// 1. Define Module Helper
func setupAuthIntegration(env *setup.TestEnvironment) usecase.AuthUseCase {
    tokenRepo := repository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB)
    userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
    // ... init other real repos ...
    
    return usecase.NewAuthUsecase(..., tokenRepo, userRepo, ...)
}

// 2. Write Granular Tests
func TestAuthIntegration_Login_Success(t *testing.T) {
    // Setup Singleton Env (Starts Docker only once)
    env := setup.SetupIntegrationEnvironment(t)
    defer env.Cleanup() // Truncates tables
    setup.CleanupDatabase(t, env.DB)

    // Init Module
    authUC := setupAuthIntegration(env)

    // Seed Data
    setup.CreateTestUser(t, env.DB, "validuser", "pass")

    // Execute Logic
    resp, token, err := authUC.Login(context.Background(), loginReq)

    // Assert
    require.NoError(t, err)
    assert.NotEmpty(t, token)
}
```

---

## 🌍 3. End-to-End (E2E) Testing (Full Flow)
**Goal:** Verify the complete HTTP request-response cycle from the client's perspective.

*   **Scope:** `tests/e2e/api/`
*   **Key Pattern:** `TestServer` wrapper with `TestClient`.

### Standard: E2E Test Structure

Use `setup.SetupTestServer(t)` which automatically connects the Gin Engine to the Singleton Integration Containers.

```go
// tests/e2e/api/auth_e2e_test.go

func TestAuthFlow_E2E(t *testing.T) {
    // 1. Start Server & Client
    server := setup.SetupTestServer(t)
    defer server.Cleanup()
    client := server.Client // Wrapper for http.Client

    // 2. Perform Request
    loginPayload := map[string]interface{}{
        "username": "admin",
        "password": "password",
    }
    resp := client.POST("/api/v1/auth/login", loginPayload)

    // 3. Assert Response
    assert.Equal(t, 200, resp.StatusCode)
    
    // 4. Chain Requests (Use Token)
    token := client.ExtractToken(resp)
    profileResp := client.GET("/api/v1/users/me", setup.WithAuth(token))
    assert.Equal(t, 200, profileResp.StatusCode)
}
```

---

## 🛡️ Security Testing Standard
Tests MUST explicitly cover security vulnerabilities. These are integrated into the main test files, not separated.

*   **SQL Injection**: `Test<Module>_Security_SQLInjection`
*   **XSS**: `Test<Module>_Security_XSS`
*   **Auth Bypass**: `Test<Module>_Security_Unauthorized`

---

## 🏃 Running Tests

```bash
# Run ALL tests (Sequential execution to prevent DB race conditions)
make test-all

# Run specific layer
make test-unit
make test-integration
make test-e2e
```