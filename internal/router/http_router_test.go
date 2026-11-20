package router

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/mocking"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access"
	accessHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key"
	apiKeyHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/delivery/http"
	apiKeyModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/model"
	apiKeyMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit"
	auditHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth"
	authHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/delivery/http"
	authMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization"
	orgHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/delivery/http"
	orgMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission"
	permHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project"
	projectHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role"
	roleHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats"
	statsHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user"
	userHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/delivery/http"
	userMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook"
	webhookHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildTestStdRouter(t *testing.T, cfg StdRouterConfig, apiKeyUC *apiKeyMocks.MockApiKeyUseCase) (http.Handler, func()) {
	t.Helper()

	logger := logrus.New()
	logger.SetOutput(io.Discard)

	authUC := authMocks.NewMockAuthUseCase(t)
	ticketMgr := mocking.NewMockTicketManager(t)
	userRepo := new(userMocks.MockUserRepository)
	orgRepo := orgMocks.NewMockOrganizationRepository(t)
	orgReader := orgMocks.NewMockIOrganizationReader(t)

	authMw := middleware.NewAuthMiddleware(authUC, logger, ticketMgr)
	apiKeyMw := middleware.NewAPIKeyMiddleware(apiKeyUC, userRepo, logger)
	tenantMw := middleware.NewTenantMiddleware(orgRepo, orgReader, logger)

	sseMgr := sse.NewManager()
	wsCtrl := &ws.WebSocketController{}

	handler := SetupStdRouter(
		cfg,
		&auth.AuthModule{AuthController: &authHttp.AuthController{}},
		&user.UserModule{UserController: &userHttp.UserController{}, UserRepo: userRepo},
		&permission.PermissionModule{PermissionController: &permHttp.PermissionController{}},
		&access.AccessModule{AccessController: &accessHttp.AccessController{}},
		&role.RoleModule{RoleController: &roleHttp.RoleController{}},
		&organization.OrganizationModule{OrganizationController: &orgHttp.OrganizationController{}},
		&audit.AuditModule{AuditController: &auditHttp.AuditController{}},
		&stats.StatsModule{StatsController: &statsHttp.StatsController{}},
		&project.ProjectModule{ProjectController: &projectHttp.ProjectController{}},
		&api_key.ApiKeyModule{Controller: &apiKeyHttp.ApiKeyController{}},
		&webhook.WebhookModule{Controller: &webhookHttp.WebhookController{}},
		authMw,
		apiKeyMw,
		nil,
		tenantMw,
		wsCtrl,
		sseMgr,
		nil,
		nil,
		nil,
		logger,
	)

	return handler, func() {
		sseMgr.Stop()
	}
}

func TestGetStdHealth(t *testing.T) {
	healthFunc := GetStdHealth(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()

	healthFunc(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"OK"`)
}

func TestBasicAuthHandler(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("metrics-data"))
	})

	authHandler := BasicAuthHandler(nextHandler, "admin", "secret123")

	t.Run("Reject without auth", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		w := httptest.NewRecorder()

		authHandler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Reject invalid credentials", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		req.SetBasicAuth("admin", "wrong")
		w := httptest.NewRecorder()

		authHandler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Allow valid credentials", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		req.SetBasicAuth("admin", "secret123")
		w := httptest.NewRecorder()

		authHandler.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "metrics-data", w.Body.String())
	})
}

func TestSetupStdRouter_RuntimeParity(t *testing.T) {
	apiKeyUC := new(apiKeyMocks.MockApiKeyUseCase)
	apiKeyUC.On("Authenticate", mock.Anything, "sk_live_session_test").Return(&apiKeyModel.ApiKeyIdentity{
		ApiKeyID:       "ak-1",
		UserID:         "user-1",
		OrganizationID: "org-1",
		Username:       "api_bot",
		Scopes:         []string{"project:view"},
	}, nil)

	cfg := StdRouterConfig{
		AllowedOrigins:   []string{"https://app.example.com"},
		MetricsEnabled:   true,
		MetricsAuth:      true,
		MetricsUser:      "prom",
		MetricsPass:      "secret",
		RateLimitEnabled: false,
	}

	h, cleanup := buildTestStdRouter(t, cfg, apiKeyUC)
	defer cleanup()

	t.Run("OPTIONS CORS preflight returns 204 instead of 405", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/api/v1/projects", nil)
		req.Header.Set("Origin", "https://app.example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("Swagger docs redirect, index.html, and doc.json", func(t *testing.T) {
		// /api/v1/docs -> 301 redirect
		reqRedirect := httptest.NewRequest(http.MethodGet, "/api/v1/docs", nil)
		wRedirect := httptest.NewRecorder()
		h.ServeHTTP(wRedirect, reqRedirect)
		assert.Equal(t, http.StatusMovedPermanently, wRedirect.Code)
		assert.Equal(t, "/api/v1/docs/index.html", wRedirect.Header().Get("Location"))

		// /api/v1/docs/ -> Swagger UI HTML
		reqRoot := httptest.NewRequest(http.MethodGet, "/api/v1/docs/", nil)
		wRoot := httptest.NewRecorder()
		h.ServeHTTP(wRoot, reqRoot)
		assert.Equal(t, http.StatusOK, wRoot.Code)
		assert.Contains(t, wRoot.Header().Get("Content-Type"), "text/html")
		assert.Contains(t, wRoot.Body.String(), "./doc.json")

		// /api/v1/docs/index.html -> Swagger UI HTML
		reqIndex := httptest.NewRequest(http.MethodGet, "/api/v1/docs/index.html", nil)
		wIndex := httptest.NewRecorder()
		h.ServeHTTP(wIndex, reqIndex)
		assert.Equal(t, http.StatusOK, wIndex.Code)
		assert.Contains(t, wIndex.Body.String(), "SwaggerUIBundle")

		// /api/v1/docs/doc.json -> valid JSON swagger doc
		reqDoc := httptest.NewRequest(http.MethodGet, "/api/v1/docs/doc.json", nil)
		wDoc := httptest.NewRecorder()
		h.ServeHTTP(wDoc, reqDoc)
		assert.Equal(t, http.StatusOK, wDoc.Code)
		assert.Contains(t, wDoc.Header().Get("Content-Type"), "application/json")

		var spec map[string]any
		require.NoError(t, json.Unmarshal(wDoc.Body.Bytes(), &spec))
		assert.Equal(t, "2.0", spec["swagger"])
	})

	t.Run("Health endpoint returns 200 and increments Prometheus totalRequests", func(t *testing.T) {
		before := middleware.GetTotalRequests()

		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		w := httptest.NewRecorder()
		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), `"status":"OK"`)
		assert.Greater(t, middleware.GetTotalRequests(), before)
	})

	t.Run("Metrics endpoint enforces BasicAuth", func(t *testing.T) {
		reqNoAuth := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		wNoAuth := httptest.NewRecorder()
		h.ServeHTTP(wNoAuth, reqNoAuth)
		assert.Equal(t, http.StatusUnauthorized, wNoAuth.Code)

		reqAuth := httptest.NewRequest(http.MethodGet, "/metrics", nil)
		reqAuth.SetBasicAuth("prom", "secret")
		wAuth := httptest.NewRecorder()
		h.ServeHTTP(wAuth, reqAuth)
		assert.Equal(t, http.StatusOK, wAuth.Code)
		assert.Contains(t, wAuth.Body.String(), "http_requests_total")
	})

	t.Run("authChain rejects API key on session-only endpoints", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
		req.Header.Set("X-API-Key", "sk_live_session_test")
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("POST /api/v1/auth/refresh enforces CSRF check", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
