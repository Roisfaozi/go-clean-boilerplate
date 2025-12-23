# E2E Testing Guide

## 📋 Overview

End-to-End (E2E) testing menguji full HTTP request-response cycle dari perspektif client dengan real server, database, dan dependencies.

---

## 🚀 Quick Start

### Run E2E Tests

```bash
# Run all E2E tests
make test-e2e

# Run specific test
go test -v ./tests/e2e/api/auth_e2e_test.go -tags=e2e

# Run with coverage
go test -coverprofile=coverage.txt ./tests/e2e/... -tags=e2e
```

---

## 🏗️ Setup Infrastructure

### Test Server Setup

```go
// ./tests/e2e/setup/test_server.go
package setup

import (
    "net/http/httptest"
    "testing"
)

type TestServer struct {
    Server   *httptest.Server
    DB       *gorm.DB
    Redis    *redis.Client
    Enforcer *casbin.Enforcer
    BaseURL  string
    Client   *TestClient
}

func SetupTestServer(t *testing.T) *TestServer {
    // Setup containers
    env := setupIntegrationEnvironment(t)
    
    // Initialize application
    cfg := &config.AppConfig{
        Server: config.ServerConfig{Port: 0},
        Database: getDatabaseConfig(env),
        Redis: getRedisConfig(env),
    }
    
    app, err := config.NewApplication(cfg)
    require.NoError(t, err)
    
    // Create test server
    server := httptest.NewServer(app.Server.Handler)
    
    client := NewTestClient(server.URL)
    
    return &TestServer{
        Server:   server,
        DB:       env.DB,
        Redis:    env.Redis,
        Enforcer: env.Enforcer,
        BaseURL:  server.URL,
        Client:   client,
    }
}

func (s *TestServer) Cleanup() {
    s.Server.Close()
    // Cleanup containers
}
```

### HTTP Client Helpers

```go
// ./tests/e2e/setup/test_client.go
package setup

type TestClient struct {
    BaseURL string
    Token   string
}

func NewTestClient(baseURL string) *TestClient {
    return &TestClient{BaseURL: baseURL}
}

func (c *TestClient) POST(path string, body interface{}, opts ...RequestOption) *Response {
    jsonBody, _ := json.Marshal(body)
    req, _ := http.NewRequest("POST", c.BaseURL+path, bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    
    for _, opt := range opts {
        opt(req)
    }
    
    resp, _ := http.DefaultClient.Do(req)
    return NewResponse(resp)
}

func (c *TestClient) GET(path string, opts ...RequestOption) *Response {
    req, _ := http.NewRequest("GET", c.BaseURL+path, nil)
    
    for _, opt := range opts {
        opt(req)
    }
    
    resp, _ := http.DefaultClient.Do(req)
    return NewResponse(resp)
}

func (c *TestClient) PUT(path string, body interface{}, opts ...RequestOption) *Response {
    jsonBody, _ := json.Marshal(body)
    req, _ := http.NewRequest("PUT", c.BaseURL+path, bytes.NewBuffer(jsonBody))
    req.Header.Set("Content-Type", "application/json")
    
    for _, opt := range opts {
        opt(req)
    }
    
    resp, _ := http.DefaultClient.Do(req)
    return NewResponse(resp)
}

func (c *TestClient) DELETE(path string, opts ...RequestOption) *Response {
    req, _ := http.NewRequest("DELETE", c.BaseURL+path, nil)
    
    for _, opt := range opts {
        opt(req)
    }
    
    resp, _ := http.DefaultClient.Do(req)
    return NewResponse(resp)
}

// Request Options
type RequestOption func(*http.Request)

func WithAuth(token string) RequestOption {
    return func(req *http.Request) {
        req.Header.Set("Authorization", "Bearer "+token)
    }
}

func WithCookie(name, value string) RequestOption {
    return func(req *http.Request) {
        req.AddCookie(&http.Cookie{Name: name, Value: value})
    }
}

func WithHeader(key, value string) RequestOption {
    return func(req *http.Request) {
        req.Header.Set(key, value)
    }
}
```

### Response Helpers

```go
// ./tests/e2e/setup/response.go
package setup

type Response struct {
    *http.Response
    BodyBytes []byte
}

func NewResponse(resp *http.Response) *Response {
    bodyBytes, _ := io.ReadAll(resp.Body)
    resp.Body.Close()
    return &Response{
        Response:  resp,
        BodyBytes: bodyBytes,
    }
}

func (r *Response) JSON(v interface{}) error {
    return json.Unmarshal(r.BodyBytes, v)
}

func (r *Response) String() string {
    return string(r.BodyBytes)
}

func (r *Response) GetCookie(name string) string {
    for _, cookie := range r.Cookies() {
        if cookie.Name == name {
            return cookie.Value
        }
    }
    return ""
}
```

---

## 🧪 Test Examples

### Authentication E2E Test

```go
// ./tests/e2e/api/auth_e2e_test.go
//go:build e2e
// +build e2e

package api

func TestAuthFlow_E2E(t *testing.T) {
    server := setup.SetupTestServer(t)
    defer server.Cleanup()
    
    client := server.Client
    
    // 1. Register
    t.Run("Register", func(t *testing.T) {
        registerReq := map[string]interface{}{
            "username": "testuser",
            "email":    "test@example.com",
            "password": "password123",
            "name":     "Test User",
        }
        
        resp := client.POST("/api/v1/users/register", registerReq)
        assert.Equal(t, 201, resp.StatusCode)
        
        var result struct {
            Data struct {
                ID       string `json:"id"`
                Username string `json:"username"`
                Email    string `json:"email"`
            } `json:"data"`
        }
        
        err := resp.JSON(&result)
        require.NoError(t, err)
        assert.NotEmpty(t, result.Data.ID)
        assert.Equal(t, "testuser", result.Data.Username)
    })
    
    // 2. Login
    var accessToken string
    var refreshToken string
    
    t.Run("Login", func(t *testing.T) {
        loginReq := map[string]interface{}{
            "username": "testuser",
            "password": "password123",
        }
        
        resp := client.POST("/api/v1/auth/login", loginReq)
        assert.Equal(t, 200, resp.StatusCode)
        
        var result struct {
            Data struct {
                AccessToken string `json:"access_token"`
                TokenType   string `json:"token_type"`
                ExpiresIn   int    `json:"expires_in"`
                User        struct {
                    ID       string `json:"id"`
                    Username string `json:"username"`
                    Role     string `json:"role"`
                } `json:"user"`
            } `json:"data"`
        }
        
        err := resp.JSON(&result)
        require.NoError(t, err)
        assert.NotEmpty(t, result.Data.AccessToken)
        assert.Equal(t, "Bearer", result.Data.TokenType)
        
        accessToken = result.Data.AccessToken
        refreshToken = resp.GetCookie("refresh_token")
    })
    
    // 3. Access Protected Endpoint
    t.Run("Access Protected Endpoint", func(t *testing.T) {
        resp := client.GET("/api/v1/users/me", WithAuth(accessToken))
        assert.Equal(t, 200, resp.StatusCode)
        
        var result struct {
            Data struct {
                ID       string `json:"id"`
                Username string `json:"username"`
            } `json:"data"`
        }
        
        err := resp.JSON(&result)
        require.NoError(t, err)
        assert.Equal(t, "testuser", result.Data.Username)
    })
    
    // 4. Refresh Token
    t.Run("Refresh Token", func(t *testing.T) {
        resp := client.POST("/api/v1/auth/refresh", nil,
            WithCookie("refresh_token", refreshToken))
        assert.Equal(t, 200, resp.StatusCode)
        
        var result struct {
            Data struct {
                AccessToken string `json:"access_token"`
            } `json:"data"`
        }
        
        err := resp.JSON(&result)
        require.NoError(t, err)
        assert.NotEmpty(t, result.Data.AccessToken)
        assert.NotEqual(t, accessToken, result.Data.AccessToken)
    })
    
    // 5. Logout
    t.Run("Logout", func(t *testing.T) {
        resp := client.POST("/api/v1/auth/logout", nil, WithAuth(accessToken))
        assert.Equal(t, 200, resp.StatusCode)
    })
    
    // 6. Verify Token Invalid After Logout
    t.Run("Token Invalid After Logout", func(t *testing.T) {
        resp := client.GET("/api/v1/users/me", WithAuth(accessToken))
        assert.Equal(t, 401, resp.StatusCode)
    })
}
```

### Authorization E2E Test

```go
// ./tests/e2e/api/rbac_e2e_test.go
func TestRBAC_Authorization_E2E(t *testing.T) {
    server := setup.SetupTestServer(t)
    defer server.Cleanup()
    
    // Create admin user
    adminToken := createUserAndLogin(t, server, "admin", "role:admin")
    
    // Create regular user
    userToken := createUserAndLogin(t, server, "user", "role:user")
    
    t.Run("Admin Can Access Admin Endpoints", func(t *testing.T) {
        resp := server.Client.GET("/api/v1/users", WithAuth(adminToken))
        assert.Equal(t, 200, resp.StatusCode)
    })
    
    t.Run("User Cannot Access Admin Endpoints", func(t *testing.T) {
        resp := server.Client.GET("/api/v1/users", WithAuth(userToken))
        assert.Equal(t, 403, resp.StatusCode)
        
        var result struct {
            Message string `json:"message"`
        }
        resp.JSON(&result)
        assert.Contains(t, result.Message, "forbidden")
    })
    
    t.Run("User Can Access Own Profile", func(t *testing.T) {
        resp := server.Client.GET("/api/v1/users/me", WithAuth(userToken))
        assert.Equal(t, 200, resp.StatusCode)
    })
    
    t.Run("Admin Can Assign Role", func(t *testing.T) {
        // Get user ID
        resp := server.Client.GET("/api/v1/users/me", WithAuth(userToken))
        var userProfile struct {
            Data struct {
                ID string `json:"id"`
            } `json:"data"`
        }
        resp.JSON(&userProfile)
        
        // Assign moderator role
        assignReq := map[string]interface{}{
            "user_id": userProfile.Data.ID,
            "role":    "role:moderator",
        }
        
        resp = server.Client.POST("/api/v1/permissions/assign-role",
            assignReq, WithAuth(adminToken))
        assert.Equal(t, 200, resp.StatusCode)
    })
}
```

### CRUD E2E Test

```go
// ./tests/e2e/api/user_e2e_test.go
func TestUserCRUD_E2E(t *testing.T) {
    server := setup.SetupTestServer(t)
    defer server.Cleanup()
    
    adminToken := createAdminAndLogin(t, server)
    
    var userID string
    
    // Create
    t.Run("Create User", func(t *testing.T) {
        createReq := map[string]interface{}{
            "username": "newuser",
            "email":    "newuser@example.com",
            "password": "password123",
            "name":     "New User",
        }
        
        resp := server.Client.POST("/api/v1/users/register", createReq)
        assert.Equal(t, 201, resp.StatusCode)
        
        var result struct {
            Data struct {
                ID string `json:"id"`
            } `json:"data"`
        }
        resp.JSON(&result)
        userID = result.Data.ID
    })
    
    // Read
    t.Run("Get User", func(t *testing.T) {
        resp := server.Client.GET("/api/v1/users/"+userID, WithAuth(adminToken))
        assert.Equal(t, 200, resp.StatusCode)
        
        var result struct {
            Data struct {
                Username string `json:"username"`
            } `json:"data"`
        }
        resp.JSON(&result)
        assert.Equal(t, "newuser", result.Data.Username)
    })
    
    // Update
    t.Run("Update User", func(t *testing.T) {
        updateReq := map[string]interface{}{
            "name": "Updated Name",
        }
        
        resp := server.Client.PUT("/api/v1/users/"+userID,
            updateReq, WithAuth(adminToken))
        assert.Equal(t, 200, resp.StatusCode)
        
        var result struct {
            Data struct {
                Name string `json:"name"`
            } `json:"data"`
        }
        resp.JSON(&result)
        assert.Equal(t, "Updated Name", result.Data.Name)
    })
    
    // Delete
    t.Run("Delete User", func(t *testing.T) {
        resp := server.Client.DELETE("/api/v1/users/"+userID, WithAuth(adminToken))
        assert.Equal(t, 200, resp.StatusCode)
    })
    
    // Verify Deleted
    t.Run("Verify User Deleted", func(t *testing.T) {
        resp := server.Client.GET("/api/v1/users/"+userID, WithAuth(adminToken))
        assert.Equal(t, 404, resp.StatusCode)
    })
}
```

### Dynamic Search E2E Test

```go
// ./tests/e2e/api/dynamic_search_e2e_test.go
func TestDynamicSearch_E2E(t *testing.T) {
    server := setup.SetupTestServer(t)
    defer server.Cleanup()
    
    adminToken := createAdminAndLogin(t, server)
    
    // Create test users
    createTestUser(t, server, "alice", "alice@example.com")
    createTestUser(t, server, "bob", "bob@example.com")
    createTestUser(t, server, "charlie", "charlie@example.com")
    
    t.Run("Search With Contains Filter", func(t *testing.T) {
        searchReq := map[string]interface{}{
            "filter": map[string]interface{}{
                "username": map[string]interface{}{
                    "type": "contains",
                    "from": "ali",
                },
            },
        }
        
        resp := server.Client.POST("/api/v1/users/search",
            searchReq, WithAuth(adminToken))
        assert.Equal(t, 200, resp.StatusCode)
        
        var result struct {
            Data []map[string]interface{} `json:"data"`
        }
        resp.JSON(&result)
        assert.Len(t, result.Data, 1)
        assert.Equal(t, "alice", result.Data[0]["username"])
    })
    
    t.Run("Search With Pagination", func(t *testing.T) {
        searchReq := map[string]interface{}{
            "page":  1,
            "limit": 2,
            "sort": map[string]string{
                "username": "asc",
            },
        }
        
        resp := server.Client.POST("/api/v1/users/search",
            searchReq, WithAuth(adminToken))
        assert.Equal(t, 200, resp.StatusCode)
        
        var result struct {
            Data   []map[string]interface{} `json:"data"`
            Paging struct {
                Page  int `json:"page"`
                Limit int `json:"limit"`
                Total int `json:"total"`
            } `json:"paging"`
        }
        resp.JSON(&result)
        assert.Len(t, result.Data, 2)
        assert.Equal(t, 1, result.Paging.Page)
    })
}
```

### Complete User Journey E2E

```go
// ./tests/e2e/workflows/complete_user_journey_test.go
func TestCompleteUserJourney_E2E(t *testing.T) {
    server := setup.SetupTestServer(t)
    defer server.Cleanup()
    
    client := server.Client
    
    // 1. Register
    registerReq := map[string]interface{}{
        "username": "journeyuser",
        "email":    "journey@example.com",
        "password": "password123",
        "name":     "Journey User",
    }
    
    resp := client.POST("/api/v1/users/register", registerReq)
    require.Equal(t, 201, resp.StatusCode)
    
    // 2. Login
    loginReq := map[string]interface{}{
        "username": "journeyuser",
        "password": "password123",
    }
    
    resp = client.POST("/api/v1/auth/login", loginReq)
    require.Equal(t, 200, resp.StatusCode)
    
    var loginResult struct {
        Data struct {
            AccessToken string `json:"access_token"`
        } `json:"data"`
    }
    resp.JSON(&loginResult)
    userToken := loginResult.Data.AccessToken
    
    // 3. Get Profile
    resp = client.GET("/api/v1/users/me", WithAuth(userToken))
    require.Equal(t, 200, resp.StatusCode)
    
    // 4. Update Profile
    updateReq := map[string]interface{}{
        "name": "Updated Journey User",
    }
    
    resp = client.PUT("/api/v1/users/me", updateReq, WithAuth(userToken))
    require.Equal(t, 200, resp.StatusCode)
    
    // 5. Admin creates role
    adminToken := createAdminAndLogin(t, server)
    
    roleReq := map[string]interface{}{
        "id":   "role:premium",
        "name": "Premium User",
    }
    
    resp = client.POST("/api/v1/roles", roleReq, WithAuth(adminToken))
    require.Equal(t, 201, resp.StatusCode)
    
    // 6. Admin assigns role to user
    var userProfile struct {
        Data struct {
            ID string `json:"id"`
        } `json:"data"`
    }
    resp = client.GET("/api/v1/users/me", WithAuth(userToken))
    resp.JSON(&userProfile)
    
    assignReq := map[string]interface{}{
        "user_id": userProfile.Data.ID,
        "role":    "role:premium",
    }
    
    resp = client.POST("/api/v1/permissions/assign-role",
        assignReq, WithAuth(adminToken))
    require.Equal(t, 200, resp.StatusCode)
    
    // 7. User accesses premium feature
    resp = client.GET("/api/v1/premium/feature", WithAuth(userToken))
    assert.Equal(t, 200, resp.StatusCode)
    
    // 8. Logout
    resp = client.POST("/api/v1/auth/logout", nil, WithAuth(userToken))
    require.Equal(t, 200, resp.StatusCode)
    
    // 9. Verify token invalid
    resp = client.GET("/api/v1/users/me", WithAuth(userToken))
    assert.Equal(t, 401, resp.StatusCode)
}
```

---

## 🔧 Best Practices

### 1. Use Helper Functions

```go
func createUserAndLogin(t *testing.T, server *TestServer, username, role string) string {
    // Create user
    registerReq := map[string]interface{}{
        "username": username,
        "email":    username + "@example.com",
        "password": "password123",
        "name":     username,
    }
    
    resp := server.Client.POST("/api/v1/users/register", registerReq)
    require.Equal(t, 201, resp.StatusCode)
    
    // Assign role if not default
    if role != "role:user" {
        var user struct {
            Data struct {
                ID string `json:"id"`
            } `json:"data"`
        }
        resp.JSON(&user)
        
        // Use admin to assign role
        // ...
    }
    
    // Login
    loginReq := map[string]interface{}{
        "username": username,
        "password": "password123",
    }
    
    resp = server.Client.POST("/api/v1/auth/login", loginReq)
    require.Equal(t, 200, resp.StatusCode)
    
    var result struct {
        Data struct {
            AccessToken string `json:"access_token"`
        } `json:"data"`
    }
    resp.JSON(&result)
    
    return result.Data.AccessToken
}
```

### 2. Test Error Scenarios

```go
func TestErrorScenarios_E2E(t *testing.T) {
    server := setup.SetupTestServer(t)
    defer server.Cleanup()
    
    t.Run("Missing Required Field", func(t *testing.T) {
        req := map[string]interface{}{
            "username": "test",
            // Missing email, password, name
        }
        
        resp := server.Client.POST("/api/v1/users/register", req)
        assert.Equal(t, 422, resp.StatusCode)
    })
    
    t.Run("Invalid Email Format", func(t *testing.T) {
        req := map[string]interface{}{
            "username": "test",
            "email":    "invalid-email",
            "password": "password123",
            "name":     "Test",
        }
        
        resp := server.Client.POST("/api/v1/users/register", req)
        assert.Equal(t, 422, resp.StatusCode)
    })
    
    t.Run("Unauthorized Access", func(t *testing.T) {
        resp := server.Client.GET("/api/v1/users")
        assert.Equal(t, 401, resp.StatusCode)
    })
    
    t.Run("Forbidden Access", func(t *testing.T) {
        userToken := createUserAndLogin(t, server, "user", "role:user")
        resp := server.Client.GET("/api/v1/users", WithAuth(userToken))
        assert.Equal(t, 403, resp.StatusCode)
    })
}
```

### 3. Clean State Between Tests

```go
func TestWithCleanState(t *testing.T) {
    server := setup.SetupTestServer(t)
    defer server.Cleanup()
    
    t.Run("Test 1", func(t *testing.T) {
        // Test logic
        
        // Cleanup
        setup.CleanupDatabase(t, server.DB)
    })
    
    t.Run("Test 2", func(t *testing.T) {
        // Fresh state
        // Test logic
    })
}
```

---

## 📊 Coverage Goals

| Feature | Target Coverage |
|---------|----------------|
| Authentication | 80% |
| Authorization | 75% |
| CRUD Operations | 75% |
| Dynamic Search | 70% |
| Real-time | 60% |
| Error Handling | 80% |

---

## 🐛 Troubleshooting

### Server Startup Issues

```go
// Add startup verification
func verifyServerStartup(t *testing.T, server *TestServer) {
    resp := server.Client.GET("/api/health")
    require.Equal(t, 200, resp.StatusCode)
}
```

### Flaky Tests

```go
// Add retry logic for flaky operations
func waitForCondition(condition func() bool, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    
    for time.Now().Before(deadline) {
        if condition() {
            return nil
        }
        time.Sleep(100 * time.Millisecond)
    }
    
    return fmt.Errorf("condition not met")
}
```

---

## 📚 References

- [httptest Package](https://pkg.go.dev/net/http/httptest)
- [Testify Documentation](https://github.com/stretchr/testify)
- [HTTP Testing Best Practices](https://go.dev/doc/tutorial/web-service-gin)
