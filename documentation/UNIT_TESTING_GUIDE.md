# Unit Testing Guide

This document outlines the standards and patterns for writing Unit Tests in the Go Clean Boilerplate project.

## 🎯 Goal
Verify business logic in isolation with **zero** external dependencies (Database, Redis, Network).

## 🛠 Libraries
- **Assertion**: `github.com/stretchr/testify/assert`
- **Mocking**: `github.com/stretchr/testify/mock`
- **Mock Generation**: `vektra/mockery`

---

## 📦 1. UseCase Testing Standard

**Pattern: Dependency Struct & Setup Helper**

To avoid cluttered test setup and improve readability, use a struct to hold all mock dependencies and a helper function to initialize them.

### Step 1: Define Dependency Struct
Create a struct in your test file to hold all necessary mocks.

```go
type userTestDeps struct {
    Repo     *mocks.MockUserRepository
    TM       *mocking.MockWithTransactionManager
    Enforcer *permMocks.IEnforcer
    AuditUC  *auditMocks.MockAuditUseCase
}
```

### Step 2: Create Setup Helper
Create a function that initializes mocks and the UseCase.

```go
func setupUserTest() (*userTestDeps, usecase.UserUseCase) {
    deps := &userTestDeps{
        Repo:     new(mocks.MockUserRepository),
        TM:       new(mocking.MockWithTransactionManager),
        Enforcer: new(permMocks.IEnforcer),
        AuditUC:  new(auditMocks.MockAuditUseCase),
    }
    
    // Suppress logs during tests
    log := logrus.New()
    log.SetOutput(io.Discard)

    // Inject dependencies
    uc := usecase.NewUserUseCase(deps.TM, log, deps.Repo, deps.Enforcer, deps.AuditUC)

    return deps, uc
}
```

### Step 3: Write Clean Tests
Use the helper to get fresh mocks for each test.

```go
func TestUserUseCase_Create_Success(t *testing.T) {
    // 1. Setup
    deps, uc := setupUserTest()

    // 2. Define Expectations
    req := &model.RegisterUserRequest{Username: "testuser"}
    deps.Repo.On("FindByUsername", mock.Anything, "testuser").Return(nil, gorm.ErrRecordNotFound)
    deps.Repo.On("Create", mock.Anything, mock.Anything).Return(nil)

    // 3. Execute
    resp, err := uc.Create(context.Background(), req)

    // 4. Assert
    assert.NoError(t, err)
    assert.NotNil(t, resp)
    deps.Repo.AssertExpectations(t)
}
```

---

## 🎮 2. Controller Testing Standard

**Pattern: HTTPTest & Factory Helper**

Controllers are tested by simulating HTTP requests and inspecting the response via `httptest.ResponseRecorder`.

### Standard Setup
Use a factory function to create the controller with mocks, and a router setup helper.

```go
// Factory for Controller with Mocks
func newTestUserController(mockUseCase *mocks.MockUserUseCase) *http.UserController {
    log := logrus.New()
    log.SetLevel(logrus.PanicLevel)
    return http.NewUserController(mockUseCase, log, validator.New())
}

// Router Setup
func setupUserTestRouter() *gin.Engine {
    gin.SetMode(gin.TestMode)
    return gin.New() // Use gin.New() to avoid default middleware noise
}
```

### Writing a Controller Test

```go
func TestUserHandler_Register_Success(t *testing.T) {
    // 1. Setup
    mockUseCase := new(mocks.MockUserUseCase)
    handler := newTestUserController(mockUseCase)
    router := setupUserTestRouter()
    router.POST("/users", handler.Register)

    // 2. Expectations
    reqBody := model.RegisterUserRequest{Username: "testuser"}
    mockUseCase.On("Create", mock.Anything, &reqBody).Return(&model.UserResponse{}, nil)

    // 3. Execute
    body, _ := json.Marshal(reqBody)
    req, _ := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    
    router.ServeHTTP(w, req)

    // 4. Assert
    assert.Equal(t, http.StatusCreated, w.Code)
    mockUseCase.AssertExpectations(t)
}
```

---

## 🗄️ 3. Repository Testing Standard

Repository tests are technically "Integration Tests" for the data layer, but they reside in the module's `test/` folder.

**Tools:**
*   **SQL Repositories**: Use `glebarez/sqlite` (In-Memory) to simulate MySQL behavior without Docker.
*   **Redis Repositories**: Use `go-redis/redismock`.

### Example: SQL Repository Test (In-Memory SQLite)

```go
func setupTestDB(t *testing.T) *gorm.DB {
    db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
    require.NoError(t, err)
    err = db.AutoMigrate(&entity.User{}) // Migrate schema
    require.NoError(t, err)
    return db
}

func TestUserRepository_Create(t *testing.T) {
    db := setupTestDB(t)
    repo := repository.NewUserRepository(db, logrus.New())

    user := &entity.User{Username: "testuser"}
    err := repo.Create(context.Background(), user)

    assert.NoError(t, err)
    
    // Verify persistence
    var stored entity.User
    db.First(&stored, "username = ?", "testuser")
    assert.Equal(t, "testuser", stored.Username)
}
```

---

## ⚡ Generating Mocks

When you modify an interface in `usecase/interface.go` or `repository/interface.go`, you MUST regenerate mocks.

```bash
make mocks
```

This command uses `.mockery.yaml` configuration to generate mocks in the appropriate folders.
