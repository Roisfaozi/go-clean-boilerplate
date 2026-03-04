package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCasbinEnforcer struct {
	mock.Mock
}

func (m *MockCasbinEnforcer) Enforce(rvals ...interface{}) (bool, error) {
	args := m.Called(rvals...)
	return args.Bool(0), args.Error(1)
}

func TestCasbinMiddleware_Authorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	c.Request = req

	// Set user info in context (user_id is the key AuthMiddleware sets)
	// For testing, we simulate that the user_id (UUID) maps to "role:admin" in Casbin via grouping policy
	// But Enforce takes the SUBJECT, which is what we pass from middleware.
	// Our middleware passes `user_id`.
	userID := "user-uuid-123"
	c.Set("user_id", userID)

	mockEnforcer := new(MockCasbinEnforcer)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	mockEnforcer.On("Enforce", userID, "global", "/api/v1/users", "GET").Return(true, nil)

	casbinMiddleware := middleware.CasbinMiddleware(mockEnforcer, logger)

	casbinMiddleware(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockEnforcer.AssertExpectations(t)
}

func TestCasbinMiddleware_Unauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/users", nil)
	c.Request = req

	userID := "user-uuid-456"
	c.Set("user_id", userID)

	mockEnforcer := new(MockCasbinEnforcer)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	mockEnforcer.On("Enforce", userID, "global", "/api/v1/users", "POST").Return(false, nil)

	casbinMiddleware := middleware.CasbinMiddleware(mockEnforcer, logger)

	casbinMiddleware(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "forbidden")
	mockEnforcer.AssertExpectations(t)
}

func TestCasbinMiddleware_EnforcerError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	c.Request = req

	userID := "user-uuid-789"
	c.Set("user_id", userID)

	mockEnforcer := new(MockCasbinEnforcer)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	mockEnforcer.On("Enforce", userID, "global", "/api/v1/users", "GET").Return(false, errors.New("casbin error"))

	casbinMiddleware := middleware.CasbinMiddleware(mockEnforcer, logger)

	casbinMiddleware(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal server error")
	mockEnforcer.AssertExpectations(t)
}

func TestCasbinMiddleware_NoRoleInContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	c.Request = req

	// No role or user-id set

	mockEnforcer := new(MockCasbinEnforcer)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	casbinMiddleware := middleware.CasbinMiddleware(mockEnforcer, logger)

	casbinMiddleware(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
	mockEnforcer.AssertNotCalled(t, "Enforce")
}

func TestCasbinMiddleware_ReleaseModeNilEnforcer(t *testing.T) {
	oldMode := gin.Mode()
	gin.SetMode(gin.ReleaseMode)
	defer gin.SetMode(oldMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/users", nil)
	c.Request = req

	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	// Pass nil enforcer in release mode
	casbinMiddleware := middleware.CasbinMiddleware(nil, logger)

	casbinMiddleware(c)

	assert.Equal(t, http.StatusOK, w.Code) // Expected to continue but log critical warning
}

func TestCasbinMiddleware_WithOrganizationContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/org/data", nil)
	c.Request = req

	userID := "user-uuid-123"
	orgID := "org-uuid-456"
	c.Set("user_id", userID)
	c.Set("organization_id", orgID)

	mockEnforcer := new(MockCasbinEnforcer)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	mockEnforcer.On("Enforce", userID, orgID, "/api/v1/org/data", "GET").Return(true, nil)

	casbinMiddleware := middleware.CasbinMiddleware(mockEnforcer, logger)

	casbinMiddleware(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockEnforcer.AssertExpectations(t)
}

func TestCasbinMiddleware_WithInvalidOrganizationContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/org/data", nil)
	c.Request = req

	userID := "user-uuid-123"
	// Set invalid type for organization_id
	c.Set("user_id", userID)
	c.Set("organization_id", 456)

	mockEnforcer := new(MockCasbinEnforcer)
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})

	// Should fallback to "global" domain
	mockEnforcer.On("Enforce", userID, "global", "/api/v1/org/data", "GET").Return(true, nil)

	casbinMiddleware := middleware.CasbinMiddleware(mockEnforcer, logger)

	casbinMiddleware(c)

	assert.Equal(t, http.StatusOK, w.Code)
	mockEnforcer.AssertExpectations(t)
}
