package middleware

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/test/mocks"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupUserStatusTest() (*gin.Engine, *mocks.MockUserRepository, *logrus.Logger) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	mockRepo := new(mocks.MockUserRepository)
	log := logrus.New()
	log.SetOutput(io.Discard)
	
	return router, mockRepo, log
}

// ============================================================================
// ✅ POSITIVE CASES
// ============================================================================

func TestUserStatusMiddleware_ActiveUser(t *testing.T) {
	router, mockRepo, log := setupUserStatusTest()
	
	userID := "user-123"
	activeUser := &entity.User{
		ID:       userID,
		Username: "activeuser",
		Status:   entity.UserStatusActive,
	}

	// Mock repository
	mockRepo.On("FindByID", mock.Anything, userID).Return(activeUser, nil)

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.Use(UserStatusMiddleware(mockRepo, log))
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	mockRepo.AssertExpectations(t)
}

// ❌ NEGATIVE CASES

func TestUserStatusMiddleware_SuspendedUser(t *testing.T) {
	router, mockRepo, log := setupUserStatusTest()
	
	userID := "user-456"
	suspendedUser := &entity.User{
		ID:       userID,
		Username: "suspendeduser",
		Status:   entity.UserStatusSuspended,
	}

	// Mock repository
	mockRepo.On("FindByID", mock.Anything, userID).Return(suspendedUser, nil)

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.Use(UserStatusMiddleware(mockRepo, log))
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "suspended")
	mockRepo.AssertExpectations(t)
}

func TestUserStatusMiddleware_BannedUser(t *testing.T) {
	router, mockRepo, log := setupUserStatusTest()
	
	userID := "user-789"
	bannedUser := &entity.User{
		ID:       userID,
		Username: "banneduser",
		Status:   entity.UserStatusBanned,
	}

	// Mock repository
	mockRepo.On("FindByID", mock.Anything, userID).Return(bannedUser, nil)

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.Use(UserStatusMiddleware(mockRepo, log))
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "banned")
	mockRepo.AssertExpectations(t)
}

func TestUserStatusMiddleware_NoUserContext(t *testing.T) {
	router, mockRepo, log := setupUserStatusTest()
	
	// Don't set user_id in context
	router.Use(UserStatusMiddleware(mockRepo, log))
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
	mockRepo.AssertNotCalled(t, "FindByID")
}

// 🔄 EDGE CASES

func TestUserStatusMiddleware_UserNotFoundInDB(t *testing.T) {
	router, mockRepo, log := setupUserStatusTest()
	
	userID := "nonexistent-user"

	// Mock repository - User not found
	mockRepo.On("FindByID", mock.Anything, userID).Return(nil, errors.New("user not found"))

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.Use(UserStatusMiddleware(mockRepo, log))
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal server error")
	mockRepo.AssertExpectations(t)
}

func TestUserStatusMiddleware_DatabaseError(t *testing.T) {
	router, mockRepo, log := setupUserStatusTest()
	
	userID := "user-101"

	// Mock repository - Database error
	mockRepo.On("FindByID", mock.Anything, userID).Return(nil, errors.New("database connection lost"))

	router.Use(func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	})
	router.Use(UserStatusMiddleware(mockRepo, log))
	
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRepo.AssertExpectations(t)
}
