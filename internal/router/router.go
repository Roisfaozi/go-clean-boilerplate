package router

import (
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access"
	accessHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit"
	auditHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth"
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
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
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
	OTEL             struct {
		Enabled     bool
		ServiceName string
	}
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

	if cfg.OTEL.Enabled {
		router.Use(otelgin.Middleware(cfg.OTEL.ServiceName))
	}

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

	// Rate Limiter Definition
	var publicLimiter, criticalLimiter, authLimiter gin.HandlerFunc

	if cfg.RateLimitEnabled {
		if cfg.RateLimitStore == "redis" {
			// Tier 1: Public API - Low limit (e.g. 10 RPS)
			publicLimiter = middleware.RateLimitMiddlewareRedis(redisClient, logger, middleware.LimiterTypeIP, 10*60, 60)
			// Tier 3: Critical Endpoints (Login) - Very Low Limit (e.g. 5 RPM)
			criticalLimiter = middleware.RateLimitMiddlewareRedis(redisClient, logger, middleware.LimiterTypeIP, 5, 60)
			// Tier 2: Authenticated User - High limit (e.g. 100 RPS)
			authLimiter = middleware.RateLimitMiddlewareRedis(redisClient, logger, middleware.LimiterTypeUser, 100*60, 60)
			logger.Info("Advanced Rate Limiter enabled: Redis store")
		} else {
			// Fallback to Memory (Global for now, as memory limiter refactor is separate task)
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
	if publicLimiter != nil {
		public.Use(publicLimiter)
	}
	{
		// Special handling for Login to use Critical Limiter
		authGroup := public.Group("/auth")
		if criticalLimiter != nil {
			authGroup.POST("/login", criticalLimiter, authModule.AuthController.Login)
		} else {
			authGroup.POST("/login", authModule.AuthController.Login)
		}

		// Other Auth Routes (Standard Public Limit)
		authGroup.POST("/refresh", authModule.AuthController.RefreshToken)
		authGroup.POST("/forgot-password", authModule.AuthController.ForgotPassword)
		authGroup.POST("/reset-password", authModule.AuthController.ResetPassword)
		authGroup.POST("/verify-email", authModule.AuthController.VerifyEmail)

		userHttp.RegisterPublicRoutes(public, userModule.UserController)
	}

	authenticated := apiV1.Group("")
	authenticated.Use(authMiddleware.ValidateToken())
	authenticated.Use(middleware.UserStatusMiddleware(userModule.UserRepo, logger))
	if authLimiter != nil {
		authenticated.Use(authLimiter)
	}
	{
		// Manually register auth routes that need authentication
		authGroup := authenticated.Group("/auth")
		authGroup.POST("/logout", authModule.AuthController.Logout)
		authGroup.POST("/resend-verification", authModule.AuthController.ResendVerification)

		userHttp.RegisterAuthenticatedRoutes(authenticated, userModule.UserController)
		permissionHttp.RegisterBatchCheckRoute(authenticated, permissionModule.PermissionController)
	}

	authorized := apiV1.Group("")
	authorized.Use(authMiddleware.ValidateToken())
	authorized.Use(middleware.UserStatusMiddleware(userModule.UserRepo, logger))
	authorized.Use(casbinMiddleware)
	if authLimiter != nil {
		authorized.Use(authLimiter)
	}
	{
		permissionHttp.RegisterPermissionRoutes(authorized, permissionModule.PermissionController)
		accessHttp.RegisterAccessRoutes(authorized, accessModule.AccessController)
		roleHttp.RegisterAuthorizedRoutes(authorized, roleModule.RoleController)
		userHttp.RegisterAuthorizedRoutes(authorized, userModule.UserController)
		auditHttp.RegisterAuthorizedRoutes(authorized, auditModule.AuditController)
	}

	return router
}
