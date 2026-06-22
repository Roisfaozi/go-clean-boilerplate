//go:build integration
// +build integration

package scenarios

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	apiKeyModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/model"
	apiKeyRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/repository"
	apiKeyUC "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/usecase"
	orgEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	orgRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	userRepo "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiKeyLifecycle_Integration(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	// Custom logger to see output
	logger := logrus.New()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(logrus.DebugLevel)

	// Repositories
	uRepo := userRepo.NewUserRepository(env.DB, logger)
	akRepo := apiKeyRepo.NewApiKeyRepository(env.DB)
	organizationRepo := orgRepo.NewOrganizationRepository(env.DB, env.Redis)

	// UseCases
	akUC := apiKeyUC.NewApiKeyUseCase(akRepo, organizationRepo, uRepo, env.Redis, logger)

	// Middlewares
	akMiddleware := middleware.NewAPIKeyMiddleware(akUC, uRepo, logger)

	ctx := context.Background()

	// 1. Create a real organization
	orgID, _ := uuid.NewV7()
	org := &orgEntity.Organization{
		ID:   orgID.String(),
		Name: "Test Org",
		Slug: "test-org",
	}
	err := env.DB.Create(org).Error
	require.NoError(t, err)

	// 2. Create a real user
	userID, _ := uuid.NewV7()
	user := &userEntity.User{
		ID:       userID.String(),
		Username: "apitestuser",
		Email:    "api@test.com",
		Status:   "active",
	}
	err = env.DB.Create(user).Error
	require.NoError(t, err)

	// 3. Generate an API Key via UseCase
	createReq := &apiKeyModel.CreateApiKeyRequest{
		Name:   "My App Key",
		Scopes: []string{"project:view"},
	}
	createRes, err := akUC.Create(ctx, user.ID, org.ID, createReq)
	require.NoError(t, err)
	require.NotEmpty(t, createRes.Key)
	apiKey := createRes.Key

	// 4. Setup Gin Router with Middleware to test the flow
	r := gin.New()
	r.Use(akMiddleware.Authenticate())
	r.GET("/protected", func(c *gin.Context) {
		uid, exists := c.Get("user_id")
		if !exists {
			c.String(http.StatusInternalServerError, "missing user_id")
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": uid})
	})
	r.GET("/scoped/projects", akMiddleware.RequireScopes("project:view"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	r.POST("/scoped/projects", akMiddleware.RequireScopes("project:manage"), func(c *gin.Context) {
		c.Status(http.StatusCreated)
	})
	r.GET("/session-only", akMiddleware.RequireUserSession(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// 5. Test Successful Access
	t.Run("Access with Valid API Key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("X-API-Key", apiKey)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, user.ID, resp["user_id"])
	})

	t.Run("Read scoped API key passes scoped read route", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/scoped/projects", nil)
		req.Header.Set("X-API-Key", apiKey)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Read scoped API key is blocked from scoped write route", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/scoped/projects", nil)
		req.Header.Set("X-API-Key", apiKey)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("API key is blocked from session-only route", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/session-only", nil)
		req.Header.Set("X-API-Key", apiKey)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	// 6. Test Invalid Key
	t.Run("Access with Invalid API Key", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("X-API-Key", "sk_live_wrong_key")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Access after Organization soft delete is blocked", func(t *testing.T) {
		err := organizationRepo.Delete(ctx, org.ID)
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("X-API-Key", apiKey)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Access after Organization restore succeeds", func(t *testing.T) {
		err := organizationRepo.Restore(ctx, org.ID)
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("X-API-Key", apiKey)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 7. Revoke Key
	t.Run("Access after Revocation", func(t *testing.T) {
		err := akUC.Revoke(ctx, org.ID, createRes.ID)
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", "/protected", nil)
		req.Header.Set("X-API-Key", apiKey)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
