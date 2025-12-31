package router

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access"
	accessHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit"
	auditHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth"
	authHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission"
	permissionHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role"
	roleHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user"
	userHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type RouterConfig struct {
	AllowedOrigins   []string
	TrustedProxies   []string
	RateLimitEnabled bool
	RateLimitRPS     float64
	RateLimitBurst   int
	RateLimitStore   string // "memory" or "redis"
}

func SetupRouter(
	cfg RouterConfig,
	authModule *auth.AuthModule,
	userModule *user.UserModule,
	permissionModule *permission.PermissionModule,
	accessModule *access.AccessModule,
	roleModule *role.RoleModule,
	auditModule *audit.AuditModule,
	authMiddleware *middleware.AuthMiddleware,
	casbinMiddleware gin.HandlerFunc,
	wsController *ws.WebSocketController,
	sseManager *sse.Manager,
	redisClient *redis.Client,
	logger *logrus.Logger,
) *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.SecurityMiddleware())
	router.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))

	if len(cfg.TrustedProxies) > 0 {
		if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
			logger.Errorf("Failed to set trusted proxies: %v", err)
		} else {
			logger.Infof("Trusted proxies set to: %v", cfg.TrustedProxies)
		}
	} else {
		// Secure default: trust no proxies if not configured.
		// However, in development (non-release), Gin might trust all or warn.
		// We explicitly set to nil (trust none) to be safe and avoid warnings.
		// NOTE: If you are behind a load balancer and don't set this, ClientIP() will return the LB IP (safe from spoofing, but rate limit will share one IP).
		if err := router.SetTrustedProxies(nil); err != nil {
			logger.Errorf("Failed to disable trusted proxies: %v", err)
		}
	}

	if cfg.RateLimitEnabled {
		if cfg.RateLimitStore == "redis" {
			// Use Redis-based rate limiter (Distributed)
			router.Use(middleware.RateLimitMiddlewareRedis(redisClient, logger, cfg.RateLimitRPS))
			logger.Info("Rate Limiter enabled: Redis store")
		} else {
			// Use In-Memory rate limiter (Local)
			router.Use(middleware.RateLimitMiddlewareMemory(cfg.RateLimitRPS, cfg.RateLimitBurst))
			logger.Info("Rate Limiter enabled: Memory store")
		}
	}

	router.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
		})
	})

	router.GET("/ws", wsController.HandleWebSocket)
	router.GET("/events", sseManager.ServeHTTP())

	apiV1 := router.Group("/api/v1")

	public := apiV1.Group("")
	{
		authHttp.RegisterPublicRoutes(public, authModule.AuthController)
		userHttp.RegisterPublicRoutes(public, userModule.UserController)
	}

	authenticated := apiV1.Group("")
	authenticated.Use(authMiddleware.ValidateToken())
	{
		authHttp.RegisterAuthenticatedRoutes(authenticated, authModule.AuthController)
	}

	authorized := apiV1.Group("")
	authorized.Use(authMiddleware.ValidateToken())
	authorized.Use(casbinMiddleware)
	{
		userHttp.RegisterAuthorizedRoutes(authorized, userModule.UserController)
		permissionHttp.RegisterPermissionRoutes(authorized, permissionModule.PermissionController)
		accessHttp.RegisterAccessRoutes(authorized, accessModule.AccessController)
		roleHttp.RegisterAuthorizedRoutes(authorized, roleModule.RoleController)
		auditHttp.RegisterAuthorizedRoutes(authorized, auditModule.AuditController)
	}

	return router
}
