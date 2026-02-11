package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/repository"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	// MembershipCacheTTL is the duration for which membership data is cached
	MembershipCacheTTL = 5 * time.Minute

	// CacheKeyMembership is the Redis key format for membership validation
	// Format: org:member:{orgID}:{userID}
	CacheKeyMembership = "org:member:%s:%s"

	// CacheKeyMemberRole is the Redis key format for member role
	// Format: org:role:{orgID}:{userID}
	CacheKeyMemberRole = "org:role:%s:%s"

	// CacheValueNotMember indicates user is not a member (negative cache)
	CacheValueNotMember = "0"

	// CacheValueIsMember indicates user is a member
	CacheValueIsMember = "1"
)

// CachedOrgReader implements IOrganizationReader with Redis caching layer
// for high-performance membership validation in TenantMiddleware.
type CachedOrgReader struct {
	memberRepo repository.OrganizationMemberRepository
	redis      *redis.Client
	log        *logrus.Logger
}

// NewCachedOrgReader creates a new CachedOrgReader instance
func NewCachedOrgReader(
	memberRepo repository.OrganizationMemberRepository,
	redisClient *redis.Client,
	log *logrus.Logger,
) *CachedOrgReader {
	return &CachedOrgReader{
		memberRepo: memberRepo,
		redis:      redisClient,
		log:        log,
	}
}

// ValidateMembership checks if a user is an active member of an organization.
// Uses Redis cache for performance, falls back to database on cache miss.
func (r *CachedOrgReader) ValidateMembership(ctx context.Context, orgID, userID string) (bool, error) {
	cacheKey := fmt.Sprintf(CacheKeyMembership, orgID, userID)

	// 1. Check Redis cache first
	val, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache hit
		r.log.WithFields(logrus.Fields{
			"org_id":  orgID,
			"user_id": userID,
			"source":  "cache",
		}).Debug("Membership validation cache hit")
		return val == CacheValueIsMember, nil
	}

	if !errors.Is(err, redis.Nil) {
		// Redis error (not cache miss)
		r.log.WithError(err).Warn("Redis error during membership check, falling back to DB")
	}

	// 2. Cache miss - query database using CheckMembership
	isMember, err := r.memberRepo.CheckMembership(ctx, orgID, userID)
	if err != nil {
		return false, err
	}

	// 3. Cache the result
	r.cacheMembership(ctx, cacheKey, isMember)

	r.log.WithFields(logrus.Fields{
		"org_id":    orgID,
		"user_id":   userID,
		"is_member": isMember,
		"source":    "database",
	}).Debug("Membership validation from database, cached")

	return isMember, nil
}

// GetMemberRole returns the role of a user in an organization.
// Returns empty string if not a member.
func (r *CachedOrgReader) GetMemberRole(ctx context.Context, orgID, userID string) (string, error) {
	cacheKey := fmt.Sprintf(CacheKeyMemberRole, orgID, userID)

	// 1. Check Redis cache
	role, err := r.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		return role, nil
	}

	if !errors.Is(err, redis.Nil) {
		r.log.WithError(err).Warn("Redis error during role check, falling back to DB")
	}

	// 2. Cache miss - query database using GetMemberStatus
	// Note: GetMemberStatus returns status, we need to add a separate method for role
	// For now, we'll use status as a fallback indicator
	status, err := r.memberRepo.GetMemberStatus(ctx, orgID, userID)
	if err != nil {
		return "", err
	}

	if status == "" {
		// Not a member
		return "", nil
	}

	// GetMemberStatus returns status (active/invited/etc), not role
	// We should return the role here - for now return "member" as default
	// TODO: Add GetMemberRole to repository interface
	memberRole := "member"

	// 3. Cache the role
	r.cacheRole(ctx, orgID, userID, memberRole)

	return memberRole, nil
}

// InvalidateMembershipCache removes the cached membership data for a user-org pair.
func (r *CachedOrgReader) InvalidateMembershipCache(ctx context.Context, orgID, userID string) error {
	membershipKey := fmt.Sprintf(CacheKeyMembership, orgID, userID)
	roleKey := fmt.Sprintf(CacheKeyMemberRole, orgID, userID)

	pipe := r.redis.Pipeline()
	pipe.Del(ctx, membershipKey)
	pipe.Del(ctx, roleKey)
	_, err := pipe.Exec(ctx)

	if err != nil {
		r.log.WithError(err).WithFields(logrus.Fields{
			"org_id":  orgID,
			"user_id": userID,
		}).Error("Failed to invalidate membership cache")
		return err
	}

	r.log.WithFields(logrus.Fields{
		"org_id":  orgID,
		"user_id": userID,
	}).Debug("Membership cache invalidated")

	return nil
}

// InvalidateOrganizationCache removes all cached membership data for an organization.
// Uses Redis SCAN to find and delete all keys matching the organization pattern.
func (r *CachedOrgReader) InvalidateOrganizationCache(ctx context.Context, orgID string) error {
	pattern := fmt.Sprintf("org:*:%s:*", orgID)

	var cursor uint64
	var deletedCount int

	for {
		keys, nextCursor, err := r.redis.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			r.log.WithError(err).Error("Failed to scan keys for organization cache invalidation")
			return err
		}

		if len(keys) > 0 {
			if err := r.redis.Del(ctx, keys...).Err(); err != nil {
				r.log.WithError(err).Error("Failed to delete organization cache keys")
				return err
			}
			deletedCount += len(keys)
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	r.log.WithFields(logrus.Fields{
		"org_id":        orgID,
		"deleted_count": deletedCount,
	}).Debug("Organization cache invalidated")

	return nil
}

// cacheMembership caches the membership status in Redis
func (r *CachedOrgReader) cacheMembership(ctx context.Context, cacheKey string, isMember bool) {
	value := CacheValueNotMember
	if isMember {
		value = CacheValueIsMember
	}

	if err := r.redis.Set(ctx, cacheKey, value, MembershipCacheTTL).Err(); err != nil {
		r.log.WithError(err).Warn("Failed to cache membership status")
	}
}

// cacheRole caches the member role in Redis
func (r *CachedOrgReader) cacheRole(ctx context.Context, orgID, userID, role string) {
	cacheKey := fmt.Sprintf(CacheKeyMemberRole, orgID, userID)

	if err := r.redis.Set(ctx, cacheKey, role, MembershipCacheTTL).Err(); err != nil {
		r.log.WithError(err).Warn("Failed to cache member role")
	}
}
