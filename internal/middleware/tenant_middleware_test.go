package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/gin-gonic/gin"
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

// MockOrganizationReader is a mock implementation of IOrganizationReader
type MockOrganizationReader struct {
	mock.Mock
}

func (m *MockOrganizationReader) ValidateMembership(ctx context.Context, orgID, userID string) (bool, error) {
	args := m.Called(ctx, orgID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockOrganizationReader) GetMemberRole(ctx context.Context, orgID, userID string) (string, error) {
	args := m.Called(ctx, orgID, userID)
	return args.String(0), args.Error(1)
}

func (m *MockOrganizationReader) InvalidateMembershipCache(ctx context.Context, orgID, userID string) error {
	args := m.Called(ctx, orgID, userID)
	return args.Error(0)
}

func (m *MockOrganizationReader) InvalidateOrganizationCache(ctx context.Context, orgID string) error {
	args := m.Called(ctx, orgID)
	return args.Error(0)
}

func setupTestRouter(middleware *TenantMiddleware) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

func TestTenantMiddleware_RequireOrganization_Success(t *testing.T) {
	// Setup mocks
	mockOrgRepo := new(MockOrganizationRepository)
	mockReader := new(MockOrganizationReader)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

	// Setup expectations
	orgID := "org-123"
	userID := "user-456"

	// Mock reader validation
	mockReader.On("ValidateMembership", mock.Anything, orgID, userID).Return(true, nil)
	mockReader.On("GetMemberRole", mock.Anything, orgID, userID).Return("admin", nil)

	// Setup router
	r := setupTestRouter(middleware)
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID) // Simulate AuthMiddleware
		c.Next()
	})
	r.Use(middleware.RequireOrganization())
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"organization_id": c.GetString("organization_id"),
			"role":            c.GetString("member_role"),
		})
	})

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(OrgIDHeader, orgID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), orgID)
	assert.Contains(t, w.Body.String(), "admin")
	mockReader.AssertExpectations(t)
}

func TestTenantMiddleware_RequireOrganization_MissingOrgHeader(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockReader := new(MockOrganizationReader)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

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
	mockReader := new(MockOrganizationReader)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

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
	mockReader := new(MockOrganizationReader)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

	orgID := "org-123"
	userID := "user-456"

	// Mock reader returns not member
	mockReader.On("ValidateMembership", mock.Anything, orgID, userID).Return(false, nil)

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
	mockReader.AssertExpectations(t)
}

func TestTenantMiddleware_RequireOrganization_SlugLookup(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockReader := new(MockOrganizationReader)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

	orgID := "org-123"
	orgSlug := "my-org"
	userID := "user-456"

	// Slug lookup returns org
	mockOrgRepo.On("FindBySlug", mock.Anything, orgSlug).Return(&entity.Organization{
		ID:   orgID,
		Slug: orgSlug,
	}, nil)
	// Membership check via reader
	mockReader.On("ValidateMembership", mock.Anything, orgID, userID).Return(true, nil)
	mockReader.On("GetMemberRole", mock.Anything, orgID, userID).Return("owner", nil)

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
	mockReader.AssertExpectations(t)
}

func TestTenantMiddleware_RequireOrganization_OrgNotFound(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockReader := new(MockOrganizationReader)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

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

func TestTenantMiddleware_RequireOrganization_Error(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockReader := new(MockOrganizationReader)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

	orgID := "org-123"
	userID := "user-456"

	// Reader returns error
	mockReader.On("ValidateMembership", mock.Anything, orgID, userID).Return(false, errors.New("reader error"))

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
	mockReader.AssertExpectations(t)
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
	mockReader := new(MockOrganizationReader)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

	orgID := "org-123"
	userID := "user-456"

	mockReader.On("InvalidateMembershipCache", mock.Anything, orgID, userID).Return(nil)

	err := middleware.InvalidateMembershipCache(context.Background(), orgID, userID)
	assert.NoError(t, err)
	mockReader.AssertExpectations(t)
}

// Unused import guard
var _ = time.Second

func TestTenantMiddleware_OptionalOrganization(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockReader := new(MockOrganizationReader)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

	t.Run("no user context", func(t *testing.T) {
		r := setupTestRouter(middleware)
		r.Use(middleware.OptionalOrganization())
		r.GET("/test", func(c *gin.Context) {
			_, exists := c.Get("organization_id")
			assert.False(t, exists)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("no org specified", func(t *testing.T) {
		r := setupTestRouter(middleware)
		r.Use(func(c *gin.Context) {
			c.Set("user_id", "user123")
			c.Next()
		})
		r.Use(middleware.OptionalOrganization())
		r.GET("/test", func(c *gin.Context) {
			_, exists := c.Get("organization_id")
			assert.False(t, exists)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("slug lookup fails", func(t *testing.T) {
		mockOrgRepo := new(MockOrganizationRepository)
		mockOrgRepo.On("FindBySlug", mock.Anything, "bad-slug").Return(nil, errors.New("not found"))
		m := NewTenantMiddleware(mockOrgRepo, mockReader, log)

		r := setupTestRouter(m)
		r.Use(func(c *gin.Context) {
			c.Set("user_id", "user123")
			c.Next()
		})
		r.Use(m.OptionalOrganization())
		r.GET("/test", func(c *gin.Context) {
			_, exists := c.Get("organization_id")
			assert.False(t, exists)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(OrgSlugHeader, "bad-slug")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("slug lookup returns nil", func(t *testing.T) {
		mockOrgRepo := new(MockOrganizationRepository)
		mockOrgRepo.On("FindBySlug", mock.Anything, "bad-slug2").Return(nil, nil)
		m := NewTenantMiddleware(mockOrgRepo, mockReader, log)

		r := setupTestRouter(m)
		r.Use(func(c *gin.Context) {
			c.Set("user_id", "user123")
			c.Next()
		})
		r.Use(m.OptionalOrganization())
		r.GET("/test", func(c *gin.Context) {
			_, exists := c.Get("organization_id")
			assert.False(t, exists)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(OrgSlugHeader, "bad-slug2")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("membership validation fails", func(t *testing.T) {
		mockReader := new(MockOrganizationReader)
		mockReader.On("ValidateMembership", mock.Anything, "org123", "user123").Return(false, errors.New("error"))
		m := NewTenantMiddleware(mockOrgRepo, mockReader, log)

		r := setupTestRouter(m)
		r.Use(func(c *gin.Context) {
			c.Set("user_id", "user123")
			c.Next()
		})
		r.Use(m.OptionalOrganization())
		r.GET("/test", func(c *gin.Context) {
			_, exists := c.Get("organization_id")
			assert.False(t, exists)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(OrgIDHeader, "org123")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("not a member", func(t *testing.T) {
		mockReader := new(MockOrganizationReader)
		mockReader.On("ValidateMembership", mock.Anything, "org123", "user123").Return(false, nil)
		m := NewTenantMiddleware(mockOrgRepo, mockReader, log)

		r := setupTestRouter(m)
		r.Use(func(c *gin.Context) {
			c.Set("user_id", "user123")
			c.Next()
		})
		r.Use(m.OptionalOrganization())
		r.GET("/test", func(c *gin.Context) {
			_, exists := c.Get("organization_id")
			assert.False(t, exists)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(OrgIDHeader, "org123")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("success with org ID", func(t *testing.T) {
		mockReader := new(MockOrganizationReader)
		mockReader.On("ValidateMembership", mock.Anything, "org123", "user123").Return(true, nil)
		mockReader.On("GetMemberRole", mock.Anything, "org123", "user123").Return("admin", nil)
		m := NewTenantMiddleware(mockOrgRepo, mockReader, log)

		r := setupTestRouter(m)
		r.Use(func(c *gin.Context) {
			c.Set("user_id", "user123")
			c.Next()
		})
		r.Use(m.OptionalOrganization())
		r.GET("/test", func(c *gin.Context) {
			orgID, _ := c.Get("organization_id")
			role, _ := c.Get("member_role")
			assert.Equal(t, "org123", orgID)
			assert.Equal(t, "admin", role)
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set(OrgIDHeader, "org123")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestTenantMiddleware_RequireOrgRole(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockReader := new(MockOrganizationReader)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

	t.Run("role not found", func(t *testing.T) {
		r := setupTestRouter(middleware)
		r.Use(middleware.RequireOrgRole("admin"))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("insufficient permissions", func(t *testing.T) {
		r := setupTestRouter(middleware)
		r.Use(func(c *gin.Context) {
			c.Set("member_role", "member")
			c.Next()
		})
		r.Use(middleware.RequireOrgRole("admin", "owner"))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("sufficient permissions", func(t *testing.T) {
		r := setupTestRouter(middleware)
		r.Use(func(c *gin.Context) {
			c.Set("member_role", "admin")
			c.Next()
		})
		r.Use(middleware.RequireOrgRole("member", "admin"))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("owner permissions (hierarchy)", func(t *testing.T) {
		r := setupTestRouter(middleware)
		r.Use(func(c *gin.Context) {
			c.Set("member_role", "owner")
			c.Next()
		})
		r.Use(middleware.RequireOrgRole("admin"))
		r.GET("/test", func(c *gin.Context) {
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestGetMemberRoleFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("exists and valid", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("member_role", "admin")

		val, ok := GetMemberRoleFromContext(c)
		assert.True(t, ok)
		assert.Equal(t, "admin", val)
	})

	t.Run("not exists", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		val, ok := GetMemberRoleFromContext(c)
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("wrong type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("member_role", 123)

		val, ok := GetMemberRoleFromContext(c)
		assert.False(t, ok)
		assert.Empty(t, val)
	})

	t.Run("empty string", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("member_role", "")

		val, ok := GetMemberRoleFromContext(c)
		assert.False(t, ok)
		assert.Empty(t, val)
	})
}

func TestTenantMiddleware_RequireOrganization_FindBySlugError(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockReader := new(MockOrganizationReader)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

	userID := "user-456"
	orgSlug := "test-org"

	mockOrgRepo.On("FindBySlug", mock.Anything, orgSlug).Return(nil, errors.New("db error"))

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

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockOrgRepo.AssertExpectations(t)
}

func TestTenantMiddleware_RequireOrganization_GetMemberRoleError(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockReader := new(MockOrganizationReader)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

	orgID := "org-123"
	userID := "user-456"

	mockReader.On("ValidateMembership", mock.Anything, orgID, userID).Return(true, nil)
	mockReader.On("GetMemberRole", mock.Anything, orgID, userID).Return("", errors.New("reader error"))

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

	assert.Equal(t, http.StatusOK, w.Code)
	mockReader.AssertExpectations(t)
}

func TestTenantMiddleware_OptionalOrganization_FindBySlugError(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockReader := new(MockOrganizationReader)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

	userID := "user-456"
	orgSlug := "test-org"

	mockOrgRepo.On("FindBySlug", mock.Anything, orgSlug).Return(nil, errors.New("db error"))

	r := setupTestRouter(middleware)
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	r.Use(middleware.OptionalOrganization())
	r.GET("/test", func(c *gin.Context) {
		_, exists := c.Get("organization_id")
		assert.False(t, exists)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(OrgSlugHeader, orgSlug)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockOrgRepo.AssertExpectations(t)
}

func TestTenantMiddleware_OptionalOrganization_FindBySlugSuccess(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockReader := new(MockOrganizationReader)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

	userID := "user-456"
	orgSlug := "test-org"
	orgID := "org-123"

	mockOrgRepo.On("FindBySlug", mock.Anything, orgSlug).Return(&entity.Organization{ID: orgID}, nil)
	mockReader.On("ValidateMembership", mock.Anything, orgID, userID).Return(true, nil)
	mockReader.On("GetMemberRole", mock.Anything, orgID, userID).Return("member", nil)

	r := setupTestRouter(middleware)
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	r.Use(middleware.OptionalOrganization())
	r.GET("/test", func(c *gin.Context) {
		id, exists := c.Get("organization_id")
		assert.True(t, exists)
		assert.Equal(t, orgID, id)
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set(OrgSlugHeader, orgSlug)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockOrgRepo.AssertExpectations(t)
	mockReader.AssertExpectations(t)
}

func TestTenantMiddleware_RequireOrganization_ParamFallbackRemoved(t *testing.T) {
	mockOrgRepo := new(MockOrganizationRepository)
	mockReader := new(MockOrganizationReader)
	log := logrus.New()

	middleware := NewTenantMiddleware(mockOrgRepo, mockReader, log)

	userID := "user-456"

	r := setupTestRouter(middleware)
	r.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	r.Use(middleware.RequireOrganization())
	r.GET("/test/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test/org-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
