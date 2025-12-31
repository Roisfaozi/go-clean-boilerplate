# Comprehensive Testing Strategy Guide

This document provides a detailed guide on the testing strategy implemented in the Go Clean Boilerplate. We use a **Testing Pyramid** approach to ensure high code quality, security, and reliability.

---

## 🏗️ 1. Unit Testing
**Goal:** Verify business logic in isolation. Fast execution, no external dependencies.

*   **Scope:** `internal/modules/*/usecase`
*   **Tools:** `testify/assert`, `mockery`
*   **Dependencies:** All external calls (DB, Redis, API) are **Mocked**.

### How to Write a Unit Test
1.  **Generate Mocks**: Run `make mocks` to update mocks in `mocks/`.
2.  **Setup Test**: Initialize the UseCase with mocked dependencies.
3.  **Define Expectations**: Use `.On("MethodName", args).Return(results)`.

### Code Example: `UserUseCase.Create`
```go
// internal/modules/user/test/use_case_test.go

func TestUserUseCase_Create_Success(t *testing.T) {
    // 1. Setup Mocks
    mockRepo := new(mocks.MockUserRepository)
    mockEnforcer := new(mocks.MockEnforcer)
    mockAudit := new(mocks.MockAuditUseCase)
    mockTM := new(MockTransactionManager) // Custom mock that runs callback immediately

    // 2. Initialize UseCase
    uc := usecase.NewUserUseCase(logger, mockTM, mockRepo, mockEnforcer, mockAudit)

    // 3. Define Input
    req := &model.RegisterUserRequest{
        Username: "testuser",
        Email:    "test@example.com",
        Password: "StrongPassword123!",
    }

    // 4. Set Expectations
    mockRepo.On("FindByUsername", mock.Anything, req.Username).Return(nil, gorm.ErrRecordNotFound)
    mockRepo.On("FindByEmail", mock.Anything, req.Email).Return(nil, gorm.ErrRecordNotFound)
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
    mockEnforcer.On("AddGroupingPolicy", mock.Anything, "role:user").Return(true, nil)
    mockAudit.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

    // 5. Execute & Assert
    resp, err := uc.Create(context.Background(), req)
    
    assert.NoError(t, err)
    assert.NotNil(t, resp)
    assert.Equal(t, req.Username, resp.Username)
    
    // 6. Verify Mocks
    mockRepo.AssertExpectations(t)
}
```

---

## 🔗 2. Integration Testing
**Goal:** Verify data integrity and interaction with real infrastructure (MySQL, Redis).

*   **Scope:** `internal/modules/*/repository`, `internal/modules/*/usecase` (integrated)
*   **Tools:** `testcontainers-go`
*   **Strategy:** **Singleton Container Pattern**. One DB/Redis instance is shared across all tests to save time. Data is cleaned via `TRUNCATE` between tests.

### How to Write an Integration Test
1.  **Setup Env**: Call `setup.SetupIntegrationEnvironment(t)`.
2.  **Seed Data**: Use `fixtures` to create necessary data (e.g., Roles).
3.  **Run Logic**: Call the Repository or UseCase directly.
4.  **Assert DB State**: Check if data is actually persisted.

### Code Example: `Auth UseCase (Login)`
```go
// tests/integration/modules/auth_integration_test.go

func TestAuthIntegration_Login_Success(t *testing.T) {
    // 1. Setup Singleton Environment (Starts Docker if not running)
    env := setup.SetupIntegrationEnvironment(t)
    defer env.Cleanup() // Truncates tables, doesn't kill container
    setup.CleanupDatabase(t, env.DB) // Ensure clean slate

    // 2. Seed Data
    user := setup.CreateTestUser(t, env.DB, "validuser", "valid@test.com", "Password123!")
    env.Enforcer.AddGroupingPolicy(user.ID, "role:user")

    // 3. Initialize Real UseCase
    authUC := usecase.NewAuthUsecase(..., env.Redis, env.DB, ...)

    // 4. Execute Logic
    req := model.LoginRequest{Username: "validuser", Password: "Password123!"}
    resp, token, err := authUC.Login(context.Background(), req)

    // 5. Assertions
    require.NoError(t, err)
    assert.NotEmpty(t, resp.AccessToken)
    
    // 6. Verify Side Effects (Redis Session)
    keys, _ := env.Redis.Keys(context.Background(), "session:*").Result()
    assert.NotEmpty(t, keys)
}
```

---

## 🌍 3. End-to-End (E2E) Testing
**Goal:** Verify the complete system flow from HTTP Request to Database.

*   **Scope:** `internal/router`, `Middleware`, `Controller`
*   **Tools:** `net/http/httptest`
*   **Strategy:** Spins up the actual Gin router connected to the Integration Test containers.

### How to Write an E2E Test
1.  **Setup Server**: Call `setup.SetupTestServer(t)`.
2.  **Create Client**: Use `server.Client` to send requests.
3.  **Assert Response**: Check HTTP Status Code and JSON Body.

### Code Example: `Protected Route Access`
```go
// tests/e2e/api/auth_e2e_test.go

func TestAuthFlow_E2E(t *testing.T) {
    server := setup.SetupTestServer(t)
    defer server.Cleanup()
    client := server.Client

    // 1. Register (Public)
    regPayload := map[string]any{
        "username": "e2euser",
        "email":    "e2e@test.com",
        "password": "Password123!",
        "fullname": "E2E Test User",
    }
    client.POST("/api/v1/users/register", regPayload)

    // 2. Login (Public)
    loginPayload := map[string]any{"username": "e2euser", "password": "Password123!"}
    resp := client.POST("/api/v1/auth/login", loginPayload)
    assert.Equal(t, 200, resp.StatusCode)
    
    // Extract Token
    var body struct { Data struct { AccessToken string `json:"access_token"` } }
    resp.JSON(&body)
    token := body.Data.AccessToken

    // 3. Get Profile (Protected)
    // Using helper WithAuth to inject Bearer token
    profileResp := client.GET("/api/v1/users/me", setup.WithAuth(token))
    
    assert.Equal(t, 200, profileResp.StatusCode)
    var profile struct { Data struct { Username string } }
    profileResp.JSON(&profile)
    assert.Equal(t, "e2euser", profile.Data.Username)
}
```

---

## 🛡️ Security Testing Scope

We explicitly test for common vulnerabilities in the **Integration** and **E2E** layers:

| Vulnerability | Testing Method | File Example |
| :--- | :--- | :--- |
| **SQL Injection** | Inject payloads like `' OR '1'='1` into login/search fields. | `auth_integration_comprehensive_test.go` |
| **XSS** | Inject `<script>` tags into profile names. | `user_integration_comprehensive_test.go` |
| **Brute Force** | Loop login attempts and verify 429/401 response. | `auth_e2e_comprehensive_test.go` |
| **Privilege Escalation** | Try accessing Admin APIs with a User token. | `rbac_e2e_test.go` |
| **Weak Passwords** | Attempt registration with short/common passwords. | `user_integration_test.go` |

---

## 🏃 Running the Tests

Use the `Makefile` for easy execution:

```bash
# 1. Run EVERYTHING (Recommended for CI)
make test-all

# 2. Run Fast Unit Tests (Dev loop)
make test-unit

# 3. Run Integration Tests (Requires Docker)
make test-integration

# 4. Run E2E Tests (Requires Docker)
make test-e2e
```
