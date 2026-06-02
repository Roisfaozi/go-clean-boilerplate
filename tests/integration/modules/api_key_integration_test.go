//go:build integration
// +build integration

package modules

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/usecase"
	orgRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiKeyIntegration_Lifecycle(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	// Initialize dependencies
	apiKeyRepo := repository.NewApiKeyRepository(env.DB)
	orgRepo := orgRepository.NewOrganizationRepository(env.DB)
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)

	uc := usecase.NewApiKeyUseCase(apiKeyRepo, orgRepo, userRepo, env.Redis, env.Logger)
	ctx := context.Background()

	// 1. Setup Test User
	user := setup.CreateTestUser(t, env.DB, "api_tester", "api@test.com", "Password123!")
	org := setup.CreateTestOrganization(t, env.DB, user.ID, "API Key Test Org", "api-key-test-org")
	orgID := org.ID

	t.Run("Create and Authenticate with Caching", func(t *testing.T) {
		// Create Key
		createReq := &model.CreateApiKeyRequest{
			Name:   "Prod Key",
			Scopes: []string{"read", "write"},
		}

		created, err := uc.Create(ctx, user.ID, orgID, createReq)
		require.NoError(t, err)
		require.NotNil(t, created)
		require.NotEmpty(t, created.Key)

		// First Authenticate (Cache Miss -> Should populate Redis)
		identity, err := uc.Authenticate(ctx, created.Key)
		require.NoError(t, err)
		assert.Equal(t, user.ID, identity.UserID)
		assert.Equal(t, "api_tester", identity.Username)

		// Verify key exists in Redis
		// Use the same hashing logic as usecase
		actualKey := created.Key[8:] // remove sk_live_
		hash := setup.HashSHA256(actualKey)
		cacheKey := fmt.Sprintf("nexusos:api_key:v1:%s", hash)

		exists := env.Redis.Exists(ctx, cacheKey).Val()
		assert.Equal(t, int64(1), exists, "Key should be cached in Redis after first authentication")

		// Second Authenticate (Cache Hit)
		// We can verify this by checking that it still works even if we "break" the DB connection briefly if needed,
		// but simple verification of identity is enough here.
		identity2, err := uc.Authenticate(ctx, created.Key)
		require.NoError(t, err)
		assert.Equal(t, identity.ApiKeyID, identity2.ApiKeyID)
	})

	t.Run("Revoke should invalidate Cache", func(t *testing.T) {
		// Create a new key
		created, _ := uc.Create(ctx, user.ID, orgID, &model.CreateApiKeyRequest{Name: "To Revoke"})

		// Authenticate to cache it
		identity, _ := uc.Authenticate(ctx, created.Key)
		actualKey := created.Key[8:]
		hash := setup.HashSHA256(actualKey)
		cacheKey := fmt.Sprintf("nexusos:api_key:v1:%s", hash)

		require.Equal(t, int64(1), env.Redis.Exists(ctx, cacheKey).Val())

		// Revoke Key
		err := uc.Revoke(ctx, orgID, identity.ApiKeyID)
		require.NoError(t, err)

		// Verify it's gone from Redis
		require.Eventually(t, func() bool {
			return env.Redis.Exists(ctx, cacheKey).Val() == 0
		}, 2*time.Second, 100*time.Millisecond, "Cache should be deleted immediately upon revocation")

		// Verify authentication fails
		_, err = uc.Authenticate(ctx, created.Key)
		assert.Error(t, err, "Authentication should fail after revocation")
	})
}
