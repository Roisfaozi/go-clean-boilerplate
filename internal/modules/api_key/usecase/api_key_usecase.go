package usecase

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/repository"
	userRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ApiKeyUseCase interface {
	Create(ctx context.Context, userID, orgID string, req *model.CreateApiKeyRequest) (*model.CreateApiKeyResponse, error)
	List(ctx context.Context, orgID string) ([]model.ApiKeyResponse, error)
	Revoke(ctx context.Context, orgID, id string) error
	Authenticate(ctx context.Context, key string) (*model.ApiKeyIdentity, error)
}

type apiKeyUseCase struct {
	repo     repository.ApiKeyRepository
	userRepo userRepository.UserRepository
	redis    *redis.Client
	log      *logrus.Logger
}

func NewApiKeyUseCase(repo repository.ApiKeyRepository, userRepo userRepository.UserRepository, redis *redis.Client, log *logrus.Logger) ApiKeyUseCase {
	return &apiKeyUseCase{
		repo:     repo,
		userRepo: userRepo,
		redis:    redis,
		log:      log,
	}
}

func (uc *apiKeyUseCase) Create(ctx context.Context, userID, orgID string, req *model.CreateApiKeyRequest) (*model.CreateApiKeyResponse, error) {
	rawKey, err := uc.generateSecureKey()
	if err != nil {
		return nil, exception.ErrInternalServer
	}

	keyHash := uc.hashKey(rawKey)

	scopesJson, _ := json.Marshal(req.Scopes)

	id, _ := uuid.NewV7()
	apiKey := &entity.ApiKey{
		ID:             id.String(),
		Name:           req.Name,
		KeyHash:        keyHash,
		OrganizationID: orgID,
		UserID:         userID,
		Scopes:         string(scopesJson),
		ExpiresAt:      req.ExpiresAt,
		IsActive:       true,
	}

	if err := uc.repo.Create(ctx, apiKey); err != nil {
		uc.log.WithFields(logrus.Fields{
			"error":  err,
			"userID": userID,
			"orgID":  orgID,
		}).Error("Failed to create API key in database")
		return nil, exception.ErrInternalServer
	}

	return &model.CreateApiKeyResponse{
		ApiKeyResponse: model.ApiKeyResponse{
			ID:             apiKey.ID,
			Name:           apiKey.Name,
			OrganizationID: apiKey.OrganizationID,
			UserID:         apiKey.UserID,
			Scopes:         req.Scopes,
			ExpiresAt:      apiKey.ExpiresAt,
			IsActive:       apiKey.IsActive,
			CreatedAt:      apiKey.CreatedAt,
		},
		Key: fmt.Sprintf("sk_live_%s", rawKey),
	}, nil
}

func (uc *apiKeyUseCase) List(ctx context.Context, orgID string) ([]model.ApiKeyResponse, error) {
	keys, err := uc.repo.ListByOrg(ctx, orgID)
	if err != nil {
		return nil, exception.ErrInternalServer
	}

	responses := make([]model.ApiKeyResponse, 0, len(keys))
	for _, k := range keys {
		var scopes []string
		_ = json.Unmarshal([]byte(k.Scopes), &scopes)

		responses = append(responses, model.ApiKeyResponse{
			ID:             k.ID,
			Name:           k.Name,
			OrganizationID: k.OrganizationID,
			UserID:         k.UserID,
			Scopes:         scopes,
			ExpiresAt:      k.ExpiresAt,
			LastUsedAt:     k.LastUsedAt,
			IsActive:       k.IsActive,
			CreatedAt:      k.CreatedAt,
		})
	}

	return responses, nil
}

func (uc *apiKeyUseCase) Revoke(ctx context.Context, orgID, id string) error {
	apiKey, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return exception.ErrNotFound
		}
		return exception.ErrInternalServer
	}

	if apiKey.OrganizationID != orgID {
		return exception.ErrForbidden
	}

	if err := uc.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate Cache
	cacheKey := fmt.Sprintf("nexusos:api_key:v1:%s", apiKey.KeyHash)
	_ = uc.redis.Del(ctx, cacheKey).Err()

	return nil
}

func (uc *apiKeyUseCase) Authenticate(ctx context.Context, key string) (*model.ApiKeyIdentity, error) {
	// Remove prefix if present
	actualKey := strings.TrimPrefix(key, "sk_live_")
	keyHash := uc.hashKey(actualKey)

	cacheKey := fmt.Sprintf("nexusos:api_key:v1:%s", keyHash)

	// Try Cache First
	if uc.redis != nil {
		val, err := uc.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			var identity model.ApiKeyIdentity
			if err := json.Unmarshal([]byte(val), &identity); err == nil {
				// Re-verify expiration
				if identity.ExpiresAt != nil && identity.ExpiresAt.Before(time.Now()) {
					return nil, exception.ErrUnauthorized
				}
				return &identity, nil
			}
		}
	}

	// Cache Miss - Query DB
	apiKey, err := uc.repo.FindByHash(ctx, keyHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrUnauthorized
		}
		return nil, exception.ErrInternalServer
	}

	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, exception.ErrUnauthorized
	}

	// Fetch User Info to complete the identity
	user, err := uc.userRepo.FindByID(ctx, apiKey.UserID)
	if err != nil {
		uc.log.WithError(err).Errorf("Failed to find user %s for API key", apiKey.UserID)
		return nil, exception.ErrUnauthorized
	}

	var scopes []string
	_ = json.Unmarshal([]byte(apiKey.Scopes), &scopes)

	identity := &model.ApiKeyIdentity{
		ApiKeyID:       apiKey.ID,
		UserID:         apiKey.UserID,
		OrganizationID: apiKey.OrganizationID,
		Username:       user.Username,
		Scopes:         scopes,
		ExpiresAt:      apiKey.ExpiresAt,
	}

	// Update last used at (Async)
	go func() {
		now := time.Now()
		apiKey.LastUsedAt = &now
		_ = uc.repo.Update(context.Background(), apiKey)
	}()

	// Save to Cache
	if uc.redis != nil {
		data, _ := json.Marshal(identity)
		_ = uc.redis.Set(ctx, cacheKey, string(data), 30*time.Minute).Err()
	}

	return identity, nil
}

func (uc *apiKeyUseCase) generateSecureKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func (uc *apiKeyUseCase) hashKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
