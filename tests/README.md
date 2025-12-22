# Testing Guide

## 📋 Overview

Proyek ini memiliki 3 jenis testing:
1. **Unit Tests** - Test untuk UseCase, Repository, Controller dengan mocks
2. **Integration Tests** - Test dengan database real (MySQL) dan Redis menggunakan testcontainers
3. **E2E Tests** - Test full HTTP request-response cycle dengan real server

## 🚀 Quick Start

### Prerequisites

```bash
# Install dependencies
go mod download

# Install testcontainers-go
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/mysql
go get github.com/testcontainers/testcontainers-go/modules/redis

# Install gjson for JSON assertions
go get github.com/tidwall/gjson
```

### Run Tests

```bash
# Unit tests only (fast, no containers)
make test

# Integration tests (requires Docker)
make test-integration

# E2E tests (requires Docker)
make test-e2e

# All tests
make test-all

# With coverage
make test-coverage-all
```

## 🏗️ Project Structure

```
./tests/
├── integration/           # Integration Tests
│   ├── setup/
│   │   ├── test_container.go    # Container setup (MySQL, Redis)
│   │   └── test_database.go     # Database helpers
│   ├── modules/
│   │   ├── auth_integration_test.go
│   │   └── user_integration_test.go
│   └── scenarios/
│       └── user_lifecycle_test.go
│
├── e2e/                   # E2E Tests
│   ├── setup/
│   │   ├── test_server.go
│   │   └── test_client.go
│   ├── api/
│   │   ├── auth_e2e_test.go
│   │   └── user_e2e_test.go
│   └── workflows/
│       └── complete_user_journey_test.go
│
├── fixtures/              # Test Data Factories
│   ├── user_factory.go
│   └── role_factory.go
│
└── helpers/               # Test Helpers
    ├── assertions.go
    └── wait.go
```

## 🧪 Integration Tests

### Running Integration Tests

```bash
# Run all integration tests
go test -v ./tests/integration/... -tags=integration

# Run specific module
go test -v ./tests/integration/modules/auth_integration_test.go -tags=integration

# Run with timeout
go test -v ./tests/integration/... -tags=integration -timeout=10m

# Run in parallel
go test -v ./tests/integration/... -tags=integration -parallel=4
```

### Writing Integration Tests

```go
//go:build integration
// +build integration

package modules

import (
    "testing"
    "github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
)

func TestMyIntegration(t *testing.T) {
    t.Parallel() // Enable parallel execution
    
    // Setup environment (MySQL + Redis containers)
    env := setup.SetupIntegrationEnvironment(t)
    defer env.Cleanup()
    
    // Clean database before test
    setup.CleanupDatabase(t, env.DB)
    
    // Your test logic here
    // ...
}
```

## 🌐 E2E Tests

### Running E2E Tests

```bash
# Run all E2E tests
go test -v ./tests/e2e/... -tags=e2e

# Run specific API test
go test -v ./tests/e2e/api/auth_e2e_test.go -tags=e2e

# Run with timeout
go test -v ./tests/e2e/... -tags=e2e -timeout=15m
```

### Writing E2E Tests

```go
//go:build e2e
// +build e2e

package api

import (
    "testing"
    "github.com/Roisfaozi/go-clean-boilerplate/tests/e2e/setup"
)

func TestMyE2E(t *testing.T) {
    // Setup test server
    server := setup.SetupTestServer(t)
    defer server.Cleanup()
    
    // Make HTTP requests
    resp := server.Client.POST("/api/v1/auth/login", loginReq)
    
    // Assertions
    assert.Equal(t, 200, resp.StatusCode)
}
```

## 🔧 Test Fixtures

### Using Factories

```go
import "github.com/Roisfaozi/go-clean-boilerplate/tests/fixtures"

// Create user with factory
userFactory := fixtures.NewUserFactory(db)
user := userFactory.Create()

// Create with overrides
admin := userFactory.Create(func(u *entity.User) {
    u.Username = "admin"
    u.Email = "admin@example.com"
})

// Create multiple users
users := userFactory.CreateMany(10)
```

## 📊 Coverage

### Generate Coverage Report

```bash
# Generate coverage for all tests
go test -coverprofile=coverage.txt -covermode=atomic -v ./... -tags=integration,e2e

# View coverage in browser
go tool cover -html=coverage.txt -o coverage.html
```

### Coverage Goals

| Layer | Target |
|-------|--------|
| Repository | 90% |
| UseCase | 85% |
| Controller | 85% |
| Integration | 70% |
| E2E | 60% |
| **Overall** | **80%** |

## 🐛 Troubleshooting

### Docker Issues

```bash
# Check Docker is running
docker ps

# Clean up old containers
docker container prune

# Check container logs
docker logs <container_id>
```

### Test Timeout

```bash
# Increase timeout
go test -timeout=20m ./tests/integration/...
```

### Port Already in Use

```bash
# Find and kill process using port
netstat -ano | findstr :3306
taskkill /PID <PID> /F
```

### Database Connection Failed

- Ensure Docker is running
- Check if MySQL container started successfully
- Wait for container to be ready (health check)
- Check logs: `docker logs <mysql_container_id>`

## 📚 Best Practices

### 1. Test Isolation
- Always clean database before test
- Use `t.Parallel()` for parallel execution
- Each test should be independent

### 2. Test Naming
```go
func TestModuleName_FunctionName_Scenario(t *testing.T)
// Example: TestUserIntegration_Create_Success
```

### 3. Cleanup
```go
func TestExample(t *testing.T) {
    env := setup.SetupIntegrationEnvironment(t)
    defer env.Cleanup() // Always cleanup
    
    // Test logic
}
```

### 4. Assertions
```go
// Use require for critical checks
require.NoError(t, err)
require.NotNil(t, result)

// Use assert for non-critical checks
assert.Equal(t, expected, actual)
assert.Contains(t, list, item)
```

## 🎯 Next Steps

1. ✅ Phase 1: Infrastructure Setup (COMPLETED)
   - Test containers
   - Database helpers
   - Fixtures & factories

2. 🔄 Phase 2: Integration Tests (IN PROGRESS)
   - Auth module ✅
   - User module ✅
   - Role module (TODO)
   - Permission module (TODO)
   - Cross-module scenarios (TODO)

3. ⏳ Phase 3: E2E Tests (PENDING)
   - Auth API
   - User API
   - RBAC enforcement
   - Complete user journey

4. ⏳ Phase 4: CI/CD Integration (PENDING)
   - Update Makefile
   - GitHub Actions workflow
   - Coverage reporting

## 📖 References

- [Testcontainers Go](https://golang.testcontainers.org/)
- [Testify](https://github.com/stretchr/testify)
- [GORM Testing](https://gorm.io/docs/testing.html)
- [Go Testing](https://go.dev/doc/tutorial/add-a-test)
