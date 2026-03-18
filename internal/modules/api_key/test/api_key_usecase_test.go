package test

import (
	"context"
	"errors"
	"io"
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
	"gorm.io/gorm"
)

type apiKeyTestDeps struct {
	Repo *mocks.MockApiKeyRepository
}

func setupApiKeyTest() (*apiKeyTestDeps, usecase.ApiKeyUseCase) {
	deps := &apiKeyTestDeps{
		Repo: new(mocks.MockApiKeyRepository),
	}
	log := logrus.New()
	log.SetOutput(io.Discard)
	uc := usecase.NewApiKeyUseCase(deps.Repo, log)
	return deps, uc
}

func TestApiKeyUseCase_Create(t *testing.T) {
	ctx := context.TODO()

	t.Run("Success", func(t *testing.T) {
		deps, uc := setupApiKeyTest()

		expiresAt := time.Now().Add(24 * time.Hour)
		req := &model.CreateApiKeyRequest{
			Name:      "Test Key",
			Scopes:    []string{"read", "write"},
			ExpiresAt: &expiresAt,
		}

		deps.Repo.EXPECT().Create(ctx, mock.AnythingOfType("*entity.ApiKey")).Return(nil).Once()

		res, err := uc.Create(ctx, "user-id", "org-id", req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res.Key)
		assert.True(t, len(res.Key) > 10)
		assert.Equal(t, "Test Key", res.Name)
		assert.Equal(t, "org-id", res.OrganizationID)
		assert.Equal(t, "user-id", res.UserID)
		assert.ElementsMatch(t, []string{"read", "write"}, res.Scopes)
		assert.Equal(t, &expiresAt, res.ExpiresAt)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		deps, uc := setupApiKeyTest()

		req := &model.CreateApiKeyRequest{
			Name: "Test Key",
		}

		deps.Repo.EXPECT().Create(ctx, mock.AnythingOfType("*entity.ApiKey")).Return(errors.New("db error")).Once()

		res, err := uc.Create(ctx, "user-id", "org-id", req)

		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
		deps.Repo.AssertExpectations(t)
	})
}

func TestApiKeyUseCase_List(t *testing.T) {
	ctx := context.TODO()

	t.Run("Success", func(t *testing.T) {
		deps, uc := setupApiKeyTest()

		keys := []*entity.ApiKey{
			{
				ID:             "key-1",
				Name:           "Key 1",
				OrganizationID: "org-id",
				UserID:         "user-id",
				Scopes:         `["read", "write"]`,
				IsActive:       true,
			},
		}

		deps.Repo.EXPECT().ListByOrg(ctx, "org-id").Return(keys, nil).Once()

		res, err := uc.List(ctx, "org-id")

		assert.NoError(t, err)
		assert.Len(t, res, 1)
		assert.Equal(t, "key-1", res[0].ID)
		assert.Equal(t, "Key 1", res[0].Name)
		assert.ElementsMatch(t, []string{"read", "write"}, res[0].Scopes)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Repository Error", func(t *testing.T) {
		deps, uc := setupApiKeyTest()

		deps.Repo.EXPECT().ListByOrg(ctx, "org-id").Return(nil, errors.New("db error")).Once()

		res, err := uc.List(ctx, "org-id")

		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
		deps.Repo.AssertExpectations(t)
	})
}

func TestApiKeyUseCase_Revoke(t *testing.T) {
	ctx := context.TODO()

	t.Run("Success", func(t *testing.T) {
		deps, uc := setupApiKeyTest()

		existingKey := &entity.ApiKey{
			ID:             "key-1",
			OrganizationID: "org-id",
		}

		deps.Repo.EXPECT().FindByID(ctx, "key-1").Return(existingKey, nil).Once()
		deps.Repo.EXPECT().Delete(ctx, "key-1").Return(nil).Once()

		err := uc.Revoke(ctx, "org-id", "key-1")

		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Key Not Found", func(t *testing.T) {
		deps, uc := setupApiKeyTest()

		deps.Repo.EXPECT().FindByID(ctx, "key-1").Return(nil, gorm.ErrRecordNotFound).Once()

		err := uc.Revoke(ctx, "org-id", "key-1")

		assert.ErrorIs(t, err, exception.ErrNotFound)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("FindByID Error", func(t *testing.T) {
		deps, uc := setupApiKeyTest()

		deps.Repo.EXPECT().FindByID(ctx, "key-1").Return(nil, errors.New("db error")).Once()

		err := uc.Revoke(ctx, "org-id", "key-1")

		assert.ErrorIs(t, err, exception.ErrInternalServer)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Forbidden - Different Org", func(t *testing.T) {
		deps, uc := setupApiKeyTest()

		existingKey := &entity.ApiKey{
			ID:             "key-1",
			OrganizationID: "different-org-id",
		}

		deps.Repo.EXPECT().FindByID(ctx, "key-1").Return(existingKey, nil).Once()

		err := uc.Revoke(ctx, "org-id", "key-1")

		assert.ErrorIs(t, err, exception.ErrForbidden)
		deps.Repo.AssertExpectations(t)
	})
}

func TestApiKeyUseCase_Authenticate(t *testing.T) {
	ctx := context.TODO()

	t.Run("Success", func(t *testing.T) {
		deps, uc := setupApiKeyTest()

		rawKey := "some-secret-key"
		prefixedKey := "sk_live_" + rawKey

		// Not checking exact hash as it's internal to the usecase, we just use mock.Anything for hash
		existingKey := &entity.ApiKey{
			ID:       "key-1",
			IsActive: true,
		}

		deps.Repo.EXPECT().FindByHash(ctx, mock.AnythingOfType("string")).Return(existingKey, nil).Once()
		deps.Repo.EXPECT().Update(ctx, mock.AnythingOfType("*entity.ApiKey")).Return(nil).Once()

		res, err := uc.Authenticate(ctx, prefixedKey)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "key-1", res.ID)
		assert.NotNil(t, res.LastUsedAt)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Key Not Found", func(t *testing.T) {
		deps, uc := setupApiKeyTest()

		deps.Repo.EXPECT().FindByHash(ctx, mock.AnythingOfType("string")).Return(nil, gorm.ErrRecordNotFound).Once()

		res, err := uc.Authenticate(ctx, "invalid-key")

		assert.ErrorIs(t, err, exception.ErrUnauthorized)
		assert.Nil(t, res)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("FindByHash Error", func(t *testing.T) {
		deps, uc := setupApiKeyTest()

		deps.Repo.EXPECT().FindByHash(ctx, mock.AnythingOfType("string")).Return(nil, errors.New("db error")).Once()

		res, err := uc.Authenticate(ctx, "invalid-key")

		assert.ErrorIs(t, err, exception.ErrInternalServer)
		assert.Nil(t, res)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Key Expired", func(t *testing.T) {
		deps, uc := setupApiKeyTest()

		pastTime := time.Now().Add(-1 * time.Hour)
		existingKey := &entity.ApiKey{
			ID:        "key-1",
			IsActive:  true,
			ExpiresAt: &pastTime,
		}

		deps.Repo.EXPECT().FindByHash(ctx, mock.AnythingOfType("string")).Return(existingKey, nil).Once()

		res, err := uc.Authenticate(ctx, "sk_live_somekey")

		assert.ErrorIs(t, err, exception.ErrUnauthorized)
		assert.Nil(t, res)
		deps.Repo.AssertExpectations(t)
	})
}
