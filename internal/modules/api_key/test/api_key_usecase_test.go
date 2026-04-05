package test

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/usecase"
	apiKeyMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/test/mocks"
	userEntity "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	userMocks "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/go-redis/redismock/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
    "gorm.io/gorm"
    "errors"
    "encoding/json"
)

type apiKeyTestDeps struct {
	Repo     *apiKeyMocks.MockApiKeyRepository
	UserRepo *userMocks.MockUserRepository
	Redis    redismock.ClientMock
}

func setupApiKeyTest() (*apiKeyTestDeps, usecase.ApiKeyUseCase) {
	redisClient, redisMock := redismock.NewClientMock()
	deps := &apiKeyTestDeps{
		Repo:     new(apiKeyMocks.MockApiKeyRepository),
		UserRepo: new(userMocks.MockUserRepository),
		Redis:    redisMock,
	}

	log := logrus.New()

	uc := usecase.NewApiKeyUseCase(deps.Repo, deps.UserRepo, redisClient, log)
	return deps, uc
}

func TestApiKeyUseCase_Create(t *testing.T) {
	deps, uc := setupApiKeyTest()
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
		req := &model.CreateApiKeyRequest{
			Name:   "Test Key",
			Scopes: []string{"read", "write"},
		}

		deps.Repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.ApiKey")).Return(nil).Once()

		res, err := uc.Create(ctx, "user1", "org1", req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.NotEmpty(t, res.Key)
		assert.Equal(t, req.Name, res.Name)
	})

	t.Run("Failure - Create", func(t *testing.T) {
		req := &model.CreateApiKeyRequest{
			Name:   "Test Key",
			Scopes: []string{"read", "write"},
		}

		deps.Repo.On("Create", mock.Anything, mock.AnythingOfType("*entity.ApiKey")).Return(errors.New("db error")).Once()

		res, err := uc.Create(ctx, "user1", "org1", req)

		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestApiKeyUseCase_List(t *testing.T) {
	deps, uc := setupApiKeyTest()
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
        keys := []*entity.ApiKey{
            {
                ID: "1",
                Name: "key1",
                Scopes: `["read"]`,
            },
        }

		deps.Repo.On("ListByOrg", mock.Anything, "org1").Return(keys, nil).Once()

		res, err := uc.List(ctx, "org1")

		assert.NoError(t, err)
		assert.NotNil(t, res)
        assert.Len(t, res, 1)
	})

	t.Run("Failure", func(t *testing.T) {
		deps.Repo.On("ListByOrg", mock.Anything, "org1").Return(nil, errors.New("db error")).Once()

		res, err := uc.List(ctx, "org1")

		assert.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestApiKeyUseCase_Revoke(t *testing.T) {
	deps, uc := setupApiKeyTest()
	ctx := context.Background()

	t.Run("Success", func(t *testing.T) {
        apiKey := &entity.ApiKey{
            ID: "1",
            OrganizationID: "org1",
            KeyHash: "hash1",
        }

		deps.Repo.On("FindByID", mock.Anything, "1").Return(apiKey, nil).Once()
		deps.Repo.On("Delete", mock.Anything, "1").Return(nil).Once()
        deps.Redis.ExpectDel("nexusos:api_key:v1:hash1").SetVal(1)

		err := uc.Revoke(ctx, "org1", "1")

		assert.NoError(t, err)
        assert.NoError(t, deps.Redis.ExpectationsWereMet())
	})

    t.Run("Failure - Not Found", func(t *testing.T) {
		deps.Repo.On("FindByID", mock.Anything, "1").Return(nil, gorm.ErrRecordNotFound).Once()

		err := uc.Revoke(ctx, "org1", "1")

		assert.Error(t, err)
	})

    t.Run("Failure - Forbidden", func(t *testing.T) {
        apiKey := &entity.ApiKey{
            ID: "1",
            OrganizationID: "org2",
            KeyHash: "hash1",
        }

		deps.Repo.On("FindByID", mock.Anything, "1").Return(apiKey, nil).Once()

		err := uc.Revoke(ctx, "org1", "1")

		assert.Error(t, err)
	})

    t.Run("Failure - Delete Error", func(t *testing.T) {
        apiKey := &entity.ApiKey{
            ID: "1",
            OrganizationID: "org1",
            KeyHash: "hash1",
        }

		deps.Repo.On("FindByID", mock.Anything, "1").Return(apiKey, nil).Once()
		deps.Repo.On("Delete", mock.Anything, "1").Return(errors.New("db error")).Once()

		err := uc.Revoke(ctx, "org1", "1")

		assert.Error(t, err)
	})
}

func TestApiKeyUseCase_Authenticate(t *testing.T) {
	deps, uc := setupApiKeyTest()
	ctx := context.Background()

    key := "sk_live_test"
    // The hash of "test" is 9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08
    hash := "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08"
    cacheKey := "nexusos:api_key:v1:" + hash

	t.Run("Success - Cache Hit", func(t *testing.T) {
        identity := model.ApiKeyIdentity{
            ApiKeyID: "1",
            UserID: "u1",
        }
        val, _ := json.Marshal(identity)
        deps.Redis.ExpectGet(cacheKey).SetVal(string(val))

		res, err := uc.Authenticate(ctx, key)

		assert.NoError(t, err)
		assert.NotNil(t, res)
        assert.Equal(t, "1", res.ApiKeyID)
        assert.NoError(t, deps.Redis.ExpectationsWereMet())
	})

	t.Run("Success - Cache Miss, DB Hit", func(t *testing.T) {
        apiKey := &entity.ApiKey{
            ID: "1",
            UserID: "u1",
            OrganizationID: "org1",
            Scopes: `["read"]`,
        }
        user := &userEntity.User{
            Username: "user1",
        }

        deps.Redis.ExpectGet(cacheKey).RedisNil()
        deps.Repo.On("FindByHash", mock.Anything, hash).Return(apiKey, nil).Once()
        deps.UserRepo.On("FindByID", mock.Anything, "u1").Return(user, nil).Once()
        deps.Repo.On("Update", mock.Anything, mock.AnythingOfType("*entity.ApiKey")).Return(nil).Once()

        // This is tricky, we can't easily mock the redis set in the go routine if we don't wait.
        // We'll just expect it to not fail.
        deps.Redis.ExpectSet(cacheKey, mock.Anything, 30*time.Minute).SetVal("OK")

		res, err := uc.Authenticate(ctx, key)

		assert.NoError(t, err)
		assert.NotNil(t, res)
	})

    t.Run("Failure - Expired in DB", func(t *testing.T) {
        now := time.Now().Add(-1 * time.Hour)
        apiKey := &entity.ApiKey{
            ID: "1",
            UserID: "u1",
            OrganizationID: "org1",
            Scopes: `["read"]`,
            ExpiresAt: &now,
        }

        deps.Redis.ExpectGet(cacheKey).RedisNil()
        deps.Repo.On("FindByHash", mock.Anything, hash).Return(apiKey, nil).Once()

		res, err := uc.Authenticate(ctx, key)

		assert.Error(t, err)
		assert.Nil(t, res)
	})

    t.Run("Failure - Not Found in DB", func(t *testing.T) {
        deps.Redis.ExpectGet(cacheKey).RedisNil()
        deps.Repo.On("FindByHash", mock.Anything, hash).Return(nil, gorm.ErrRecordNotFound).Once()

		res, err := uc.Authenticate(ctx, key)

		assert.Error(t, err)
		assert.Nil(t, res)
	})

    t.Run("Failure - User Not Found", func(t *testing.T) {
        apiKey := &entity.ApiKey{
            ID: "1",
            UserID: "u1",
            OrganizationID: "org1",
            Scopes: `["read"]`,
        }

        deps.Redis.ExpectGet(cacheKey).RedisNil()
        deps.Repo.On("FindByHash", mock.Anything, hash).Return(apiKey, nil).Once()
        deps.UserRepo.On("FindByID", mock.Anything, "u1").Return(nil, errors.New("db err")).Once()

		res, err := uc.Authenticate(ctx, key)

		assert.Error(t, err)
		assert.Nil(t, res)
	})
}
