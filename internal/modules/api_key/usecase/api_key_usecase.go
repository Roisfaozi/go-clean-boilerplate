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
	orgRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
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

const (
	cacheKeyOrganizationStatus = "nexusos:org_status:%s"
	organizationStatusActive   = "active"
	organizationStatusDeleted  = "deleted"
	organizationStatusCacheTTL = 30 * time.Second
)

type apiKeyUseCase struct {
	repo     repository.ApiKeyRepository
	orgRepo  orgRepository.OrganizationRepository
	userRepo userRepository.UserRepository
	redis    *redis.Client
	log      *logrus.Logger
}

func NewApiKeyUseCase(repo repository.ApiKeyRepository, orgRepo orgRepository.OrganizationRepository, userRepo userRepository.UserRepository, redis *redis.Client, log *logrus.Logger) ApiKeyUseCase {
	return &apiKeyUseCase{
		repo:     repo,
		orgRepo:  orgRepo,
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
				if err := uc.ensureOrganizationAccessible(ctx, identity.OrganizationID); err != nil {
					return nil, err
				}
				return &identity, nil
			}
		}
	}

	// Cache Miss - Query DB
	apiKey, err := uc.repo.FindByHash(ctx, keyHash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			uc.log.Warn("API key not found during authentication")
			return nil, exception.ErrUnauthorized
		}
		uc.log.WithError(err).Error("Failed to load API key during authentication")
		return nil, exception.ErrInternalServer
	}

	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, exception.ErrUnauthorized
	}

	if err := uc.ensureOrganizationAccessible(ctx, apiKey.OrganizationID); err != nil {
		return nil, err
	}

	// Fetch User Info to complete the identity
	user, err := uc.userRepo.FindByID(ctx, apiKey.UserID)
	if err != nil {
		uc.log.WithError(err).WithFields(logrus.Fields{
			"user_id":         apiKey.UserID,
			"organization_id": apiKey.OrganizationID,
			"api_key_id":      apiKey.ID,
		}).Error("Failed to find user for API key authentication")
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

func (uc *apiKeyUseCase) ensureOrganizationAccessible(ctx context.Context, orgID string) error {
	if orgID == "" || uc.orgRepo == nil {
		return nil
	}

	cacheKey := fmt.Sprintf(cacheKeyOrganizationStatus, orgID)
	if uc.redis != nil {
		status, err := uc.redis.Get(ctx, cacheKey).Result()
		if err == nil {
			switch status {
			case organizationStatusActive:
				return nil
			case organizationStatusDeleted:
				uc.log.WithField("organization_id", orgID).Warn("Organization is soft-deleted for API key authentication")
				return exception.ErrUnauthorized
			}
		}

		if !errors.Is(err, redis.Nil) {
			uc.log.WithError(err).Warn("Redis error reading organization status cache")
		}
	}

	org, err := uc.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		uc.log.WithError(err).WithField("organization_id", orgID).Error("Failed to validate organization for API key")
		return exception.ErrInternalServer
	}
	if org == nil {
		uc.cacheOrganizationStatus(ctx, orgID, organizationStatusDeleted)
		uc.log.WithField("organization_id", orgID).Warn("Organization not accessible for API key authentication")
		return exception.ErrUnauthorized
	}

	uc.cacheOrganizationStatus(ctx, orgID, organizationStatusActive)
	return nil
}

func (uc *apiKeyUseCase) cacheOrganizationStatus(ctx context.Context, orgID, status string) {
	if uc.redis == nil {
		return
	}

	cacheKey := fmt.Sprintf(cacheKeyOrganizationStatus, orgID)
	if err := uc.redis.Set(ctx, cacheKey, status, organizationStatusCacheTTL).Err(); err != nil {
		uc.log.WithError(err).WithField("organization_id", orgID).Warn("Failed to cache organization status")
	}
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
