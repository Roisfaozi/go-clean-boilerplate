package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Helper to create a dummy router for testing
func createTestRouter(cfg RouterConfig) *gin.Engine {
	// We pass nil for most modules as we are testing router config/middleware here, not business logic
	return SetupRouter(
		cfg,
		&auth.AuthModule{}, // Dummy modules
		&user.UserModule{},
		&permission.PermissionModule{},
		&access.AccessModule{},
		&role.RoleModule{},
		&audit.AuditModule{},
		&middleware.AuthMiddleware{},
		func(c *gin.Context) { c.Next() }, // Dummy Casbin Middleware
		&ws.WebSocketController{},
		sse.NewManager(),
		nil,            // db
		&redis.Client{}, // redisClient
		logrus.New(),
	)
}

func TestTrustedProxies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Should not trust X-Forwarded-For by default", func(t *testing.T) {
		cfg := RouterConfig{
			AllowedOrigins: []string{"*"},
			// TrustedProxies empty
		}

		router := createTestRouter(cfg)
		router.GET("/test-ip", func(c *gin.Context) {
			c.String(200, c.ClientIP())
		})

		req, _ := http.NewRequest("GET", "/test-ip", nil)
		// Spoofed IP
		req.Header.Set("X-Forwarded-For", "1.2.3.4")
		req.RemoteAddr = "10.0.0.1:12345" // Real IP (e.g. Load Balancer)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return RemoteAddr (10.0.0.1) because we don't trust proxies
		assert.Equal(t, "10.0.0.1", w.Body.String())
	})

	t.Run("Should trust configured proxy", func(t *testing.T) {
		cfg := RouterConfig{
			AllowedOrigins: []string{"*"},
			TrustedProxies: []string{"10.0.0.1"},
		}

		router := createTestRouter(cfg)
		router.GET("/test-ip-trusted", func(c *gin.Context) {
			c.String(200, c.ClientIP())
		})

		req, _ := http.NewRequest("GET", "/test-ip-trusted", nil)
		// Spoofed IP (or Real Client IP forwarded by LB)
		req.Header.Set("X-Forwarded-For", "1.2.3.4")
		req.RemoteAddr = "10.0.0.1:12345" // Real IP matches trusted proxy

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return X-Forwarded-For (1.2.3.4) because we trust 10.0.0.1
		assert.Equal(t, "1.2.3.4", w.Body.String())
	})
}
