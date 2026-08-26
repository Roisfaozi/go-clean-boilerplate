package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/test/mocks"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	"github.com/go-redis/redismock/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLogger() *logrus.Logger {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	return log
}

func TestCachedOrgReader_ValidateMembership_CacheHit(t *testing.T) {
	// Arrange
	db, mockClient := redismock.NewClientMock()
	mockRepo := mocks.NewMockOrganizationMemberRepository(t)
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"
	cacheKey := "org:member:org-123:user-456"

	// Mock Redis GET returning cached "1" (is member)
	mockClient.ExpectGet(cacheKey).SetVal("1")

	// Act
	isMember, err := reader.ValidateMembership(ctx, orgID, userID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, isMember)
	assert.NoError(t, mockClient.ExpectationsWereMet())
}

func TestCachedOrgReader_ValidateMembership_CacheHit_NotMember(t *testing.T) {
	// Arrange
	db, mockClient := redismock.NewClientMock()
	mockRepo := mocks.NewMockOrganizationMemberRepository(t)
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"
	cacheKey := "org:member:org-123:user-456"

	// Mock Redis GET returning cached "0" (not member)
	mockClient.ExpectGet(cacheKey).SetVal("0")

	// Act
	isMember, err := reader.ValidateMembership(ctx, orgID, userID)

	// Assert
	assert.NoError(t, err)
	assert.False(t, isMember)
	assert.NoError(t, mockClient.ExpectationsWereMet())
}

func TestCachedOrgReader_ValidateMembership_CacheMiss_IsMember(t *testing.T) {
	// Arrange
	db, mockClient := redismock.NewClientMock()

	mockRepo := mocks.NewMockOrganizationMemberRepository(t)
	mockRepo.On("CheckMembership", mock.Anything, "org-123", "user-456").Return(true, nil)
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"
	membershipKey := "org:member:org-123:user-456"

	// Mock Redis cache miss, then SET calls
	mockClient.ExpectGet(membershipKey).RedisNil()
	mockClient.ExpectSet(membershipKey, "1", 5*time.Minute).SetVal("OK")

	// Act
	isMember, err := reader.ValidateMembership(ctx, orgID, userID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, isMember)
	assert.NoError(t, mockClient.ExpectationsWereMet())
}

func TestCachedOrgReader_ValidateMembership_CacheMiss_NotMember(t *testing.T) {
	// Arrange
	db, mockClient := redismock.NewClientMock()

	mockRepo := mocks.NewMockOrganizationMemberRepository(t)
	mockRepo.On("CheckMembership", mock.Anything, "org-123", "user-456").Return(false, nil)
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"
	membershipKey := "org:member:org-123:user-456"

	// Mock Redis cache miss, then SET calls
	mockClient.ExpectGet(membershipKey).RedisNil()
	mockClient.ExpectSet(membershipKey, "0", 5*time.Minute).SetVal("OK")

	// Act
	isMember, err := reader.ValidateMembership(ctx, orgID, userID)

	// Assert
	assert.NoError(t, err)
	assert.False(t, isMember)
	assert.NoError(t, mockClient.ExpectationsWereMet())
}

func TestCachedOrgReader_ValidateMembership_DBError(t *testing.T) {
	// Arrange
	db, mockClient := redismock.NewClientMock()

	dbErr := errors.New("database connection error")
	mockRepo := mocks.NewMockOrganizationMemberRepository(t)
	mockRepo.On("CheckMembership", mock.Anything, "org-123", "user-456").Return(false, dbErr)
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"
	membershipKey := "org:member:org-123:user-456"

	// Mock Redis cache miss
	mockClient.ExpectGet(membershipKey).RedisNil()

	// Act
	isMember, err := reader.ValidateMembership(ctx, orgID, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, dbErr, err)
	assert.False(t, isMember)
}

func TestCachedOrgReader_GetMemberRole_CacheHit(t *testing.T) {
	// Arrange
	db, mockClient := redismock.NewClientMock()
	mockRepo := mocks.NewMockOrganizationMemberRepository(t)
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"
	roleKey := "org:role:v2:org-123:user-456"

	// Mock Redis GET returning cached role
	mockClient.ExpectGet(roleKey).SetVal("admin")

	// Act
	role, err := reader.GetMemberRole(ctx, orgID, userID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "admin", role)
	assert.NoError(t, mockClient.ExpectationsWereMet())
}

func TestCachedOrgReader_GetMemberRole_CacheMiss(t *testing.T) {
	// Arrange
	db, mockClient := redismock.NewClientMock()
	mockRepo := mocks.NewMockOrganizationMemberRepository(t)
	mockRepo.On("GetMemberRoleName", mock.Anything, "org-123", "user-456").Return("member", nil)
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"
	roleKey := "org:role:v2:org-123:user-456"

	// Mock Redis cache miss, then SET
	mockClient.ExpectGet(roleKey).RedisNil()
	mockClient.ExpectSet(roleKey, "member", usecase.MembershipCacheTTL).SetVal("OK")

	// Act
	role, err := reader.GetMemberRole(ctx, orgID, userID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "member", role)
	assert.NoError(t, mockClient.ExpectationsWereMet())
}

func TestCachedOrgReader_InvalidateMembershipCache(t *testing.T) {
	// Arrange
	db, mockClient := redismock.NewClientMock()
	mockRepo := mocks.NewMockOrganizationMemberRepository(t)
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"
	membershipKey := "org:member:org-123:user-456"
	roleKey := "org:role:v2:org-123:user-456"

	// Mock Redis pipeline DEL
	mockClient.ExpectDel(membershipKey).SetVal(1)
	mockClient.ExpectDel(roleKey).SetVal(1)

	// Act
	err := reader.InvalidateMembershipCache(ctx, orgID, userID)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mockClient.ExpectationsWereMet())
}

func TestCachedOrgReader_InvalidateOrganizationCache(t *testing.T) {
	// Arrange
	db, mockClient := redismock.NewClientMock()
	mockRepo := mocks.NewMockOrganizationMemberRepository(t)
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	pattern := "org:*:org-123:*"
	statusKey := "nexusos:org_status:org-123"

	// Mock Redis SCAN - returns keys and cursor 0 (stop)
	mockClient.ExpectScan(0, pattern, 100).SetVal([]string{"key1", "key2"}, 0)

	// Mock Redis DEL for the found keys and status key
	mockClient.ExpectDel("key1", "key2").SetVal(2)
	mockClient.ExpectDel(statusKey).SetVal(1)

	// Act
	err := reader.InvalidateOrganizationCache(ctx, orgID)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mockClient.ExpectationsWereMet())
}
