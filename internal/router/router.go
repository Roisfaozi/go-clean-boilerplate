package router

import (
	"net/http"

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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

type RouterConfig struct {
	AllowedOrigins   []string
	TrustedProxies   []string
	RateLimitEnabled bool
	RateLimitRPS     float64
	RateLimitBurst   int
	RateLimitStore   string
	MetricsEnabled   bool
	MetricsAuth      bool
	MetricsUser      string
	MetricsPass      string
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
	db *gorm.DB,
	redisClient *redis.Client,
	logger *logrus.Logger,
) *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.RequestIDMiddleware())

	if cfg.MetricsEnabled {
		router.Use(middleware.PrometheusMiddleware())
	}
	router.GET("/ws", wsController.HandleWebSocket)
	router.GET("/events", sseManager.ServeHTTP())

	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.SecurityMiddleware())
	router.Use(middleware.CORSMiddleware(cfg.AllowedOrigins))

	if len(cfg.TrustedProxies) > 0 {
		if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
			logger.Fatalf("Failed to set trusted proxies (invalid CIDR?): %v", err)
		} else {
			logger.Infof("Trusted proxies set to: %v", cfg.TrustedProxies)
		}
	} else {
		if err := router.SetTrustedProxies(nil); err != nil {
			logger.Fatalf("Failed to disable trusted proxies: %v", err)
		}
	}

	if cfg.RateLimitEnabled {
		if cfg.RateLimitStore == "redis" {
			router.Use(middleware.RateLimitMiddlewareRedis(redisClient, logger, cfg.RateLimitRPS))
			logger.Info("Rate Limiter enabled: Redis store")
		} else {
			router.Use(middleware.RateLimitMiddlewareMemory(cfg.RateLimitRPS, cfg.RateLimitBurst))
			logger.Info("Rate Limiter enabled: Memory store")
		}
	}

	router.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/api/health", func(c *gin.Context) {
		status := "OK"
		details := make(map[string]string)

		if db != nil {
			sqlDB, err := db.DB()
			if err != nil {
				status = "DEGRADED"
				details["mysql"] = "CONNECTION_ERROR"
			} else if err := sqlDB.Ping(); err != nil {
				status = "DEGRADED"
				details["mysql"] = "DOWN"
			} else {
				details["mysql"] = "UP"
			}
		}

		if redisClient != nil {
			if err := redisClient.Ping(c.Request.Context()).Err(); err != nil {
				status = "DEGRADED"
				details["redis"] = "DOWN"
			} else {
				details["redis"] = "UP"
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  status,
			"details": details,
		})
	})

	if cfg.MetricsEnabled {
		metricsGroup := router.Group("/metrics")
		if cfg.MetricsAuth {
			metricsGroup.Use(gin.BasicAuth(gin.Accounts{
				cfg.MetricsUser: cfg.MetricsPass,
			}))
		}
		metricsGroup.GET("", gin.WrapH(promhttp.Handler()))
	}

	apiV1 := router.Group("/api/v1")

	public := apiV1.Group("")
	{
		authHttp.RegisterPublicRoutes(public, authModule.AuthController)
		userHttp.RegisterPublicRoutes(public, userModule.UserController)
	}

	authenticated := apiV1.Group("")
	authenticated.Use(authMiddleware.ValidateToken())
	authenticated.Use(middleware.UserStatusMiddleware(userModule.UserRepo, logger))
	{
		authHttp.RegisterAuthenticatedRoutes(authenticated, authModule.AuthController)
		userHttp.RegisterAuthenticatedRoutes(authenticated, userModule.UserController)
		permissionHttp.RegisterBatchCheckRoute(authenticated, permissionModule.PermissionController)
	}

	authorized := apiV1.Group("")
	authorized.Use(authMiddleware.ValidateToken())
	authorized.Use(middleware.UserStatusMiddleware(userModule.UserRepo, logger))
	authorized.Use(casbinMiddleware)
	{
		permissionHttp.RegisterPermissionRoutes(authorized, permissionModule.PermissionController)
		accessHttp.RegisterAccessRoutes(authorized, accessModule.AccessController)
		roleHttp.RegisterAuthorizedRoutes(authorized, roleModule.RoleController)
		userHttp.RegisterAuthorizedRoutes(authorized, userModule.UserController)
		auditHttp.RegisterAuthorizedRoutes(authorized, auditModule.AuditController)
	}

	return router
}
