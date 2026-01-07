# E2E Testing Guide

## 📋 Overview

End-to-End (E2E) testing verifies the complete HTTP request-response cycle from the client's perspective using a real server instance, database, and Redis cache.

---

## 🚀 Quick Start

### Run E2E Tests

```bash
# Run all E2E tests
make test-e2e

# Run specific test file
go test -v ./tests/e2e/api/auth_e2e_test.go -tags=e2e

# Run with coverage report
go test -coverprofile=coverage_e2e.out ./tests/e2e/... -tags=e2e
```

---

## 🏗️ Infrastructure Setup

### Test Server (`tests/e2e/setup/test_server.go`)
We use `httptest.Server` to spin up the actual Gin engine. It connects to the **Singleton Containers** (MySQL/Redis) provided by the integration setup.

```go
func SetupTestServer(t *testing.T) *TestServer {
    // 1. Get Singleton DB/Redis containers
    env := integrationSetup.SetupIntegrationEnvironment(t)
    
    // 2. Initialize App Config with container addresses
    cfg := &config.AppConfig{
        // ... config pointing to test containers ...
    }
    
    // 3. Start App & Test Server
    app, _ := config.NewApplication(cfg)
    server := httptest.NewServer(app.Server.Handler)
    
    return &TestServer{...}
}
```

### Test Client (`tests/e2e/setup/test_client.go`)
A custom wrapper around `http.Client` to simplify JSON requests and Auth headers.

```go
client.POST("/api/v1/auth/login", loginReq)
client.GET("/api/v1/users/me", setup.WithAuth(token))
```

---

## 🧪 Test Examples

### 1. Authentication Flow

```go
func TestAuthFlow_E2E(t *testing.T) {
    server := setup.SetupTestServer(t)
    defer server.Cleanup()
    client := server.Client
    
    // 1. Register
    registerReq := map[string]any{
        "username": "testuser",
        "email":    "test@example.com",
        "password": "SecurePass123!",
        "fullname": "Test User",
    }
    resp := client.POST("/api/v1/users/register", registerReq)
    assert.Equal(t, 201, resp.StatusCode)
    
    // 2. Login
    loginReq := map[string]any{
        "username": "testuser",
        "password": "SecurePass123!",
    }
    resp = client.POST("/api/v1/auth/login", loginReq)
    assert.Equal(t, 200, resp.StatusCode)
    
    // Extract Token
    var result struct {
        Data struct {
            AccessToken string `json:"access_token"`
        } `json:"data"`
    }
    resp.JSON(&result)
    token := result.Data.AccessToken
    
    // 3. Access Protected Route
    resp = client.GET("/api/v1/users/me", setup.WithAuth(token))
    assert.Equal(t, 200, resp.StatusCode)
}
```

### 2. RBAC Authorization

```go
func TestRBAC_E2E(t *testing.T) {
    server := setup.SetupTestServer(t)
    // ... helper to create admin & user ...
    
    t.Run("User Cannot Access Admin Route", func(t *testing.T) {
        resp := client.GET("/api/v1/users", setup.WithAuth(userToken))
        assert.Equal(t, 403, resp.StatusCode)
    })
    
    t.Run("Admin Can Access Admin Route", func(t *testing.T) {
        resp := client.GET("/api/v1/users", setup.WithAuth(adminToken))
        assert.Equal(t, 200, resp.StatusCode)
    })
}
```

---

## 🔧 Best Practices

1.  **Use Helper Functions**: Create helpers like `createUserAndLogin(t, server)` to avoid repetition in complex flows.
2.  **Verify Status Codes**: Always assert `resp.StatusCode` first.
3.  **Check Response Body**: For success cases, verify returned data structure. For errors, verify the error message.
4.  **Sequential Execution**: E2E tests run in the same process against shared containers. Avoid `t.Parallel()` to prevent data collision.

## 🐛 Troubleshooting

| Issue | Potential Cause |
| :--- | :--- |
| **401 Unauthorized** | Token expired (check `test_server.go` config) or Redis session missing. |
| **403 Forbidden** | User role lacks Casbin policy for the endpoint/method. |
| **422 Unprocessable** | Invalid request payload (e.g., incorrect JSON field name, weak password). |
| **Connection Refused** | Docker containers not running or healthy. |
