package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockOrganizationRepository is a mock implementation of OrganizationRepository
type MockOrganizationRepository struct {
	mock.Mock
}

func (m *MockOrganizationRepository) Create(ctx context.Context, org *entity.Organization, ownerRoleID string) error {
	args := m.Called(ctx, org, ownerRoleID)
	return args.Error(0)
}

func (m *MockOrganizationRepository) FindByID(ctx context.Context, id string) (*entity.Organization, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) FindBySlug(ctx context.Context, slug string) (*entity.Organization, error) {
	args := m.Called(ctx, slug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) SlugExists(ctx context.Context, slug string) (bool, error) {
	args := m.Called(ctx, slug)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrganizationRepository) FindUserOrganizations(ctx context.Context, userID string) ([]*entity.Organization, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.Organization), args.Error(1)
}

func (m *MockOrganizationRepository) Update(ctx context.Context, org *entity.Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockOrganizationRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockOrganizationMemberRepository is a mock implementation of OrganizationMemberRepository
type MockOrganizationMemberRepository struct {
	mock.Mock
}

func (m *MockOrganizationMemberRepository) CheckMembership(ctx context.Context, orgID, userID string) (bool, error) {
	args := m.Called(ctx, orgID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrganizationMemberRepository) GetMemberStatus(ctx context.Context, orgID, userID string) (string, error) {
	args := m.Called(ctx, orgID, userID)
	return args.String(0), args.Error(1)
}

func (m *MockOrganizationMemberRepository) AddMember(ctx context.Context, member *entity.OrganizationMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockOrganizationMemberRepository) RemoveMember(ctx context.Context, orgID, userID string) error {
	args := m.Called(ctx, orgID, userID)
	return args.Error(0)
}

func (m *MockOrganizationMemberRepository) UpdateMemberRole(ctx context.Context, orgID, userID, roleID string) error {
	args := m.Called(ctx, orgID, userID, roleID)
	return args.Error(0)
}

func (m *MockOrganizationMemberRepository) UpdateMemberStatus(ctx context.Context, orgID, userID, status string) error {
	args := m.Called(ctx, orgID, userID, status)
	return args.Error(0)
}

func (m *MockOrganizationMemberRepository) FindMembers(ctx context.Context, orgID string) ([]*entity.OrganizationMember, error) {
	args := m.Called(ctx, orgID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entity.OrganizationMember), args.Error(1)
}

func setupTestRouter(middleware *TenantMiddleware) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestTenantMiddleware_RequireOrganization_Success(t *testing.T) {
	// Setup mocks
	mockOrgRepo := new(MockOrganizationRepository)
	mockMemberRepo := new(MockOrganizationMemberRepository)
	redisClient, redisMock := redismock.NewClientMock()
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockMemberRepo, redisClient, log)

	// Setup expectations
	orgID := "org-123"
	userID := "user-456"
	cacheKey := MembershipCachePrefix + orgID + ":" + userID

	// Redis cache miss
	redisMock.ExpectGet(cacheKey).RedisNil()
	// DB lookup returns active member
	mockMemberRepo.On("GetMemberStatus", mock.Anything, orgID, userID).Return("active", nil)
	// Cache the result
	membership := &MembershipCache{OrgID: orgID, Status: "active"}
	membershipJSON, _ := json.Marshal(membership)
	redisMock.ExpectSet(cacheKey, membershipJSON, MembershipCacheTTL).SetVal("OK")

	// Setup router
	r := setupTestRouter(middleware)
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID) // Simulate AuthMiddleware
		c.Next()
	})
	r.Use(middleware.RequireOrganization())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"organization_id": c.GetString("organization_id")})
	})

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(OrgIDHeader, orgID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), orgID)
	mockMemberRepo.AssertExpectations(t)
}

func TestTenantMiddleware_RequireOrganization_MissingOrgHeader(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockMemberRepo := new(MockOrganizationMemberRepository)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockMemberRepo, nil, log)

	r := setupTestRouter(middleware)
	r.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Next()
	})
	r.Use(middleware.RequireOrganization())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	// No org header set
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTenantMiddleware_RequireOrganization_NotAuthenticated(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockMemberRepo := new(MockOrganizationMemberRepository)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockMemberRepo, nil, log)

	r := setupTestRouter(middleware)
	// No auth middleware - user_id not set
	r.Use(middleware.RequireOrganization())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(OrgIDHeader, "org-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTenantMiddleware_RequireOrganization_NotMember(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockMemberRepo := new(MockOrganizationMemberRepository)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockMemberRepo, nil, log)

	orgID := "org-123"
	userID := "user-456"

	// DB lookup returns empty (not a member)
	mockMemberRepo.On("GetMemberStatus", mock.Anything, orgID, userID).Return("", nil)

	r := setupTestRouter(middleware)
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	r.Use(middleware.RequireOrganization())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(OrgIDHeader, orgID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockMemberRepo.AssertExpectations(t)
}

func TestTenantMiddleware_RequireOrganization_BannedMember(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockMemberRepo := new(MockOrganizationMemberRepository)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockMemberRepo, nil, log)

	orgID := "org-123"
	userID := "user-456"

	// DB lookup returns banned status
	mockMemberRepo.On("GetMemberStatus", mock.Anything, orgID, userID).Return("banned", nil)

	r := setupTestRouter(middleware)
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	r.Use(middleware.RequireOrganization())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(OrgIDHeader, orgID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	mockMemberRepo.AssertExpectations(t)
}

func TestTenantMiddleware_RequireOrganization_SlugLookup(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockMemberRepo := new(MockOrganizationMemberRepository)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockMemberRepo, nil, log)

	orgID := "org-123"
	orgSlug := "my-org"
	userID := "user-456"

	// Slug lookup returns org
	mockOrgRepo.On("FindBySlug", mock.Anything, orgSlug).Return(&entity.Organization{
		ID:   orgID,
		Slug: orgSlug,
	}, nil)
	// Membership check
	mockMemberRepo.On("GetMemberStatus", mock.Anything, orgID, userID).Return("active", nil)

	r := setupTestRouter(middleware)
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	r.Use(middleware.RequireOrganization())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"organization_id": c.GetString("organization_id")})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(OrgSlugHeader, orgSlug) // Use slug instead of ID
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), orgID)
	mockOrgRepo.AssertExpectations(t)
	mockMemberRepo.AssertExpectations(t)
}

func TestTenantMiddleware_RequireOrganization_OrgNotFound(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockMemberRepo := new(MockOrganizationMemberRepository)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockMemberRepo, nil, log)

	orgSlug := "non-existent-org"
	userID := "user-456"

	// Slug lookup returns nil (not found)
	mockOrgRepo.On("FindBySlug", mock.Anything, orgSlug).Return(nil, nil)

	r := setupTestRouter(middleware)
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	r.Use(middleware.RequireOrganization())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(OrgSlugHeader, orgSlug)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockOrgRepo.AssertExpectations(t)
}

func TestTenantMiddleware_RequireOrganization_CacheHit(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockMemberRepo := new(MockOrganizationMemberRepository)
	redisClient, redisMock := redismock.NewClientMock()
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockMemberRepo, redisClient, log)

	orgID := "org-123"
	userID := "user-456"
	cacheKey := MembershipCachePrefix + orgID + ":" + userID

	// Redis returns cached membership
	membership := &MembershipCache{OrgID: orgID, Status: "active"}
	membershipJSON, _ := json.Marshal(membership)
	redisMock.ExpectGet(cacheKey).SetVal(string(membershipJSON))

	// No DB call expected!

	r := setupTestRouter(middleware)
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	r.Use(middleware.RequireOrganization())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"organization_id": c.GetString("organization_id")})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(OrgIDHeader, orgID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Verify no DB call was made
	mockMemberRepo.AssertNotCalled(t, "GetMemberStatus")
}

func TestTenantMiddleware_RequireOrganization_DBError(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockMemberRepo := new(MockOrganizationMemberRepository)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockMemberRepo, nil, log)

	orgID := "org-123"
	userID := "user-456"

	// DB returns error
	mockMemberRepo.On("GetMemberStatus", mock.Anything, orgID, userID).Return("", errors.New("database error"))

	r := setupTestRouter(middleware)
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	r.Use(middleware.RequireOrganization())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(OrgIDHeader, orgID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockMemberRepo.AssertExpectations(t)
}

func TestGetOrganizationIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("exists and valid", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("organization_id", "org-123")

		orgID, ok := GetOrganizationIDFromContext(c)
		assert.True(t, ok)
		assert.Equal(t, "org-123", orgID)
	})

	t.Run("not exists", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		orgID, ok := GetOrganizationIDFromContext(c)
		assert.False(t, ok)
		assert.Empty(t, orgID)
	})

	t.Run("wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("organization_id", 123) // int instead of string

		orgID, ok := GetOrganizationIDFromContext(c)
		assert.False(t, ok)
		assert.Empty(t, orgID)
	})
}

func TestInvalidateMembershipCache(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockMemberRepo := new(MockOrganizationMemberRepository)
	redisClient, redisMock := redismock.NewClientMock()
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockMemberRepo, redisClient, log)

	orgID := "org-123"
	userID := "user-456"
	cacheKey := MembershipCachePrefix + orgID + ":" + userID

	redisMock.ExpectDel(cacheKey).SetVal(1)

	err := middleware.InvalidateMembershipCache(context.Background(), orgID, userID)
	assert.NoError(t, err)
}

// Unused import guard
var _ = time.Second
