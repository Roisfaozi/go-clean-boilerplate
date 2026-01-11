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

func createTestRouter(cfg RouterConfig) *gin.Engine {
	return SetupRouter(
		cfg,
		&auth.AuthModule{},
		&user.UserModule{},
		&permission.PermissionModule{},
		&access.AccessModule{},
		&role.RoleModule{},
		&audit.AuditModule{},
		&middleware.AuthMiddleware{},
		func(c *gin.Context) { c.Next() },
		&ws.WebSocketController{},
		sse.NewManager(),
		nil,
		&redis.Client{},
		logrus.New(),
	)
}

func TestTrustedProxies(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Should not trust X-Forwarded-For by default", func(t *testing.T) {
		cfg := RouterConfig{
			AllowedOrigins: []string{"*"},
		}

		router := createTestRouter(cfg)
		router.GET("/test-ip", func(c *gin.Context) {
			c.String(200, c.ClientIP())
		})

		req, _ := http.NewRequest("GET", "/test-ip", nil)

		req.Header.Set("X-Forwarded-For", "1.2.3.4")
		req.RemoteAddr = "10.0.0.1:12345"

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

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

		req.Header.Set("X-Forwarded-For", "1.2.3.4")
		req.RemoteAddr = "10.0.0.1:12345"

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, "1.2.3.4", w.Body.String())
	})
}
