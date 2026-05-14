package test

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestApiKeyUseCase_Create(t *testing.T) {
	repo := new(mocks.MockApiKeyRepository)
	log := logrus.New()
	uc := usecase.NewApiKeyUseCase(repo, log)

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
	log := logrus.New()
	uc := usecase.NewApiKeyUseCase(repo, log)

	ctx := context.Background()
	rawKey := "some-secure-key"
	fullKey := "sk_live_" + rawKey

	// We need to know the hash to mock FindByHash correctly
	// But since hashKey is private, we can either make it public or
	// just mock Anything if we trust the logic.

	apiKey := &entity.ApiKey{
		ID:             "key-1",
		UserID:         "user-1",
		OrganizationID: "org-1",
		IsActive:       true,
	}

	repo.On("FindByHash", ctx, mock.Anything).Return(apiKey, nil)
	repo.On("Update", ctx, mock.Anything).Return(nil)

	res, err := uc.Authenticate(ctx, fullKey)

	assert.NoError(t, err)
	assert.Equal(t, apiKey.ID, res.ID)
	repo.AssertExpectations(t)
}

func TestApiKeyUseCase_Authenticate_Expired(t *testing.T) {
	repo := new(mocks.MockApiKeyRepository)
	log := logrus.New()
	uc := usecase.NewApiKeyUseCase(repo, log)

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
