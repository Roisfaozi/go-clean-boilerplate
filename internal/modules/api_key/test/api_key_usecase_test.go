package test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/usecase"
	orgMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	userMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestApiKeyUseCase_Create(t *testing.T) {
	repo := new(mocks.MockApiKeyRepository)
	log := logrus.New()
	uc := usecase.NewApiKeyUseCase(repo, nil, nil, nil, log)

	ctx := context.Background()
	userID := "user-1"
	orgID := "org-1"
	req := &model.CreateApiKeyRequest{
		Name:   "Test Key",
		Scopes: []string{"read"},
	}

	repo.On("Create", ctx, mock.AnythingOfType("*entity.ApiKey")).Return(nil)

	res, err := uc.Create(ctx, userID, orgID, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "Test Key", res.Name)
	assert.Contains(t, res.Key, "sk_live_")
	repo.AssertExpectations(t)
}

func TestApiKeyUseCase_Authenticate(t *testing.T) {
	repo := new(mocks.MockApiKeyRepository)
	userRepo := new(userMocks.MockUserRepository)
	log := logrus.New()
	uc := usecase.NewApiKeyUseCase(repo, nil, userRepo, nil, log)

	ctx := context.Background()
	rawKey := "some-secure-key"
	fullKey := "sk_live_" + rawKey

	apiKey := &entity.ApiKey{
		ID:             "key-1",
		UserID:         "user-1",
		OrganizationID: "org-1",
		IsActive:       true,
	}

	user := &userEntity.User{
		ID:       "user-1",
		Username: "testuser",
	}

	repo.On("FindByHash", ctx, mock.Anything).Return(apiKey, nil)
	userRepo.On("FindByID", ctx, "user-1").Return(user, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)

	res, err := uc.Authenticate(ctx, fullKey)

	assert.NoError(t, err)
	assert.Equal(t, apiKey.ID, res.ApiKeyID)
	assert.Equal(t, user.Username, res.Username)

	// Wait for async update
	time.Sleep(100 * time.Millisecond)
	repo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
}

func TestApiKeyUseCase_Authenticate_CachedOrganizationStatus(t *testing.T) {
	srv, err := miniredis.Run()
	if err != nil {
		t.Skipf("miniredis unavailable in current environment: %v", err)
	}
	defer srv.Close()

	redisClient := redis.NewClient(&redis.Options{Addr: srv.Addr()})
	log := logrus.New()
	orgRepo := new(orgMocks.MockOrganizationRepository)
	uc := usecase.NewApiKeyUseCase(new(mocks.MockApiKeyRepository), orgRepo, nil, redisClient, log)

	ctx := context.Background()
	rawKey := "cached-key"
	fullKey := "sk_live_" + rawKey
	hash := sha256.Sum256([]byte(rawKey))
	cacheKey := fmt.Sprintf("nexusos:api_key:v1:%s", hex.EncodeToString(hash[:]))
	statusKey := fmt.Sprintf("nexusos:org_status:%s", "org-1")

	identity := &model.ApiKeyIdentity{
		ApiKeyID:       "key-1",
		UserID:         "user-1",
		OrganizationID: "org-1",
		Username:       "testuser",
		Scopes:         []string{"read"},
	}

	data, err := json.Marshal(identity)
	require.NoError(t, err)

	err = redisClient.Set(ctx, cacheKey, string(data), 30*time.Minute).Err()
	require.NoError(t, err)
	err = redisClient.Set(ctx, statusKey, "active", 30*time.Second).Err()
	require.NoError(t, err)

	res, err := uc.Authenticate(ctx, fullKey)

	assert.NoError(t, err)
	assert.Equal(t, identity.ApiKeyID, res.ApiKeyID)
	assert.Equal(t, identity.Username, res.Username)
	orgRepo.AssertNotCalled(t, "FindByID", mock.Anything, mock.Anything)
}

func TestApiKeyUseCase_Authenticate_Expired(t *testing.T) {
	repo := new(mocks.MockApiKeyRepository)
	log := logrus.New()
	uc := usecase.NewApiKeyUseCase(repo, nil, nil, nil, log)

	ctx := context.Background()
	past := time.Now().Add(-1 * time.Hour)
	apiKey := &entity.ApiKey{
		ID:        "key-1",
		ExpiresAt: &past,
		IsActive:  true,
	}

	repo.On("FindByHash", ctx, mock.Anything).Return(apiKey, nil)

	res, err := uc.Authenticate(ctx, "sk_live_any")

	assert.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, exception.ErrUnauthorized, err)
}
