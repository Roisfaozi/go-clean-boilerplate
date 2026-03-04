package test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/usecase"
	"github.com/go-redis/redismock/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// MockMemberRepository is a simple mock for IOrganizationMemberRepository
type MockMemberRepository struct {
	CheckMembershipFunc func(ctx context.Context, orgID, userID string) (bool, error)
	GetMemberStatusFunc func(ctx context.Context, orgID, userID string) (string, error)
	GetMemberRoleFunc   func(ctx context.Context, orgID, userID string) (string, error)
}

func (m *MockMemberRepository) CheckMembership(ctx context.Context, orgID, userID string) (bool, error) {
	if m.CheckMembershipFunc != nil {
		return m.CheckMembershipFunc(ctx, orgID, userID)
	}
	return false, nil
}

func (m *MockMemberRepository) GetMemberStatus(ctx context.Context, orgID, userID string) (string, error) {
	if m.GetMemberStatusFunc != nil {
		return m.GetMemberStatusFunc(ctx, orgID, userID)
	}
	return "", nil
}

func (m *MockMemberRepository) GetMemberRole(ctx context.Context, orgID, userID string) (string, error) {
	if m.GetMemberRoleFunc != nil {
		return m.GetMemberRoleFunc(ctx, orgID, userID)
	}
	return "", nil
}

// Methods required by interface but not used in these tests
func (m *MockMemberRepository) AddMember(ctx context.Context, member *entity.OrganizationMember) error {
	return nil
}

func (m *MockMemberRepository) RemoveMember(ctx context.Context, orgID, userID string) error {
	return nil
}

func (m *MockMemberRepository) UpdateMemberRole(ctx context.Context, orgID, userID, roleID string) error {
	return nil
}

func (m *MockMemberRepository) UpdateMemberStatus(ctx context.Context, orgID, userID, status string) error {
	return nil
}

func (m *MockMemberRepository) FindMembers(ctx context.Context, orgID string) ([]*entity.OrganizationMember, error) {
	return nil, nil
}

func newTestLogger() *logrus.Logger {
	log := logrus.New()
	log.SetLevel(logrus.DebugLevel)
	return log
}

func TestCachedOrgReader_ValidateMembership_CacheHit(t *testing.T) {
	// Arrange
	db, mock := redismock.NewClientMock()
	mockRepo := &MockMemberRepository{}
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"
	cacheKey := "org:member:org-123:user-456"

	// Mock Redis GET returning cached "1" (is member)
	mock.ExpectGet(cacheKey).SetVal("1")

	// Act
	isMember, err := reader.ValidateMembership(ctx, orgID, userID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, isMember)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCachedOrgReader_ValidateMembership_CacheHit_NotMember(t *testing.T) {
	// Arrange
	db, mock := redismock.NewClientMock()
	mockRepo := &MockMemberRepository{}
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"
	cacheKey := "org:member:org-123:user-456"

	// Mock Redis GET returning cached "0" (not member)
	mock.ExpectGet(cacheKey).SetVal("0")

	// Act
	isMember, err := reader.ValidateMembership(ctx, orgID, userID)

	// Assert
	assert.NoError(t, err)
	assert.False(t, isMember)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCachedOrgReader_ValidateMembership_CacheMiss_IsMember(t *testing.T) {
	// Arrange
	db, mock := redismock.NewClientMock()

	mockRepo := &MockMemberRepository{
		CheckMembershipFunc: func(ctx context.Context, orgID, userID string) (bool, error) {
			return true, nil
		},
	}
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"
	membershipKey := "org:member:org-123:user-456"

	// Mock Redis cache miss, then SET calls
	mock.ExpectGet(membershipKey).RedisNil()
	mock.ExpectSet(membershipKey, "1", 5*time.Minute).SetVal("OK")

	// Act
	isMember, err := reader.ValidateMembership(ctx, orgID, userID)

	// Assert
	assert.NoError(t, err)
	assert.True(t, isMember)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCachedOrgReader_ValidateMembership_CacheMiss_NotMember(t *testing.T) {
	// Arrange
	db, mock := redismock.NewClientMock()

	mockRepo := &MockMemberRepository{
		CheckMembershipFunc: func(ctx context.Context, orgID, userID string) (bool, error) {
			return false, nil
		},
	}
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"
	membershipKey := "org:member:org-123:user-456"

	// Mock Redis cache miss, then SET for negative result
	mock.ExpectGet(membershipKey).RedisNil()
	mock.ExpectSet(membershipKey, "0", 5*time.Minute).SetVal("OK")

	// Act
	isMember, err := reader.ValidateMembership(ctx, orgID, userID)

	// Assert
	assert.NoError(t, err)
	assert.False(t, isMember)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCachedOrgReader_ValidateMembership_DBError(t *testing.T) {
	// Arrange
	db, mock := redismock.NewClientMock()

	dbErr := errors.New("database connection error")
	mockRepo := &MockMemberRepository{
		CheckMembershipFunc: func(ctx context.Context, orgID, userID string) (bool, error) {
			return false, dbErr
		},
	}
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"
	membershipKey := "org:member:org-123:user-456"

	// Mock Redis cache miss
	mock.ExpectGet(membershipKey).RedisNil()

	// Act
	isMember, err := reader.ValidateMembership(ctx, orgID, userID)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, dbErr, err)
	assert.False(t, isMember)
}

func TestCachedOrgReader_GetMemberRole_CacheHit(t *testing.T) {
	// Arrange
	db, mock := redismock.NewClientMock()
	mockRepo := &MockMemberRepository{}
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"
	roleKey := "org:role:org-123:user-456"

	// Mock Redis GET returning cached role
	mock.ExpectGet(roleKey).SetVal("admin")

	// Act
	role, err := reader.GetMemberRole(ctx, orgID, userID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "admin", role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCachedOrgReader_GetMemberRole_CacheMiss(t *testing.T) {
	// Arrange
	db, mock := redismock.NewClientMock()
	mockRepo := &MockMemberRepository{
		GetMemberRoleFunc: func(ctx context.Context, orgID, userID string) (string, error) {
			return "member", nil
		},
	}
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"
	roleKey := "org:role:org-123:user-456"

	// Mock Redis cache miss, then SET
	mock.ExpectGet(roleKey).RedisNil()
	mock.ExpectSet(roleKey, "member", usecase.MembershipCacheTTL).SetVal("OK")

	// Act
	role, err := reader.GetMemberRole(ctx, orgID, userID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "member", role)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCachedOrgReader_InvalidateMembershipCache(t *testing.T) {
	// Arrange
	db, mock := redismock.NewClientMock()
	mockRepo := &MockMemberRepository{}
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	userID := "user-456"
	membershipKey := "org:member:org-123:user-456"
	roleKey := "org:role:org-123:user-456"

	// Mock Redis pipeline DEL
	mock.ExpectDel(membershipKey).SetVal(1)
	mock.ExpectDel(roleKey).SetVal(1)

	// Act
	err := reader.InvalidateMembershipCache(ctx, orgID, userID)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCachedOrgReader_InvalidateOrganizationCache(t *testing.T) {
	// Arrange
	db, mock := redismock.NewClientMock()
	mockRepo := &MockMemberRepository{}
	log := newTestLogger()

	reader := usecase.NewCachedOrgReader(mockRepo, db, log)

	ctx := context.Background()
	orgID := "org-123"
	pattern := "org:*:org-123:*"

	// Mock Redis SCAN - returns keys and cursor 0 (stop)
	mock.ExpectScan(0, pattern, 100).SetVal([]string{"key1", "key2"}, 0)

	// Mock Redis DEL for the found keys
	mock.ExpectDel("key1", "key2").SetVal(2)

	// Act
	err := reader.InvalidateOrganizationCache(ctx, orgID)

	// Assert
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
