package router

import (
	"net/http"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access"
	accessHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key"
	api_keyHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/api_key/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit"
	auditHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization"
	organizationHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission"
	permissionHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/project"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role"
	roleHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/stats"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user"
	userHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook"
	webhookHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/webhook/delivery/http"
	_ "github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/tus/tusd/v2/pkg/handler"
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
	organizationModule *organization.OrganizationModule,
	auditModule *audit.AuditModule,
	statsModule *stats.StatsModule,
	projectModule *project.ProjectModule,
	apiKeyModule *api_key.ApiKeyModule,
	webhookModule *webhook.WebhookModule,
	authMiddleware *middleware.AuthMiddleware,
	apiKeyMiddleware *middleware.APIKeyMiddleware,
	casbinMiddleware gin.HandlerFunc,
	tenantMiddleware *middleware.TenantMiddleware,
	wsController *ws.WebSocketController,
	sseManager *sse.Manager,
	db *gorm.DB,
	redisClient *redis.Client,
	tusHandler *handler.Handler,
	logger *logrus.Logger,
) *gin.Engine {
	router := gin.New()
	router.RedirectTrailingSlash = true
	router.RedirectFixedPath = true

	if cfg.OTEL.Enabled {
		router.Use(otelgin.Middleware(cfg.OTEL.ServiceName))
	}

	router.Use(middleware.RequestIDMiddleware())

	if cfg.MetricsEnabled {
		router.Use(middleware.PrometheusMiddleware())
	}

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
	var idempotencyMiddleware gin.HandlerFunc

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

	if redisClient != nil {
		idempotencyMiddleware = middleware.Idempotency(redisClient, logger, 24*time.Hour)
	} else {
		idempotencyMiddleware = func(c *gin.Context) { c.Next() }
	}

	apiV1 := router.Group("/api/v1")
	apiV1.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	if cfg.MetricsEnabled {

		metricsGroup := router.Group("/metrics")
		if cfg.MetricsAuth {
			metricsGroup.Use(gin.BasicAuth(gin.Accounts{
				cfg.MetricsUser: cfg.MetricsPass,
			}))
		}
		metricsGroup.GET("", gin.WrapH(promhttp.Handler()))
	}

	apiV1.GET("/events", authMiddleware.ValidateToken(), sseManager.GinServeHTTP())
	apiV1.GET("/ws", authMiddleware.ValidateWebSocketToken(), wsController.HandleWebSocket)
	apiV1.GET("/health", GetHealth(db, redisClient))

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
		authGroup.POST("/refresh", middleware.CSRFMiddleware(logger), idempotencyMiddleware, authModule.AuthController.RefreshToken)
		authGroup.POST("/forgot-password", idempotencyMiddleware, authModule.AuthController.ForgotPassword)
		authGroup.POST("/reset-password", idempotencyMiddleware, authModule.AuthController.ResetPassword)
		authGroup.POST("/verify-email", idempotencyMiddleware, authModule.AuthController.VerifyEmail)
		authGroup.POST("/register", idempotencyMiddleware, authModule.AuthController.Register)
		authGroup.GET("/sso/:provider", authModule.AuthController.SSOLogin)
		authGroup.GET("/sso/:provider/callback", authModule.AuthController.SSOCallback)

		userHttp.RegisterPublicRoutes(public, userModule.UserController)
		organizationHttp.RegisterPublicRoutes(public, organizationModule.OrganizationController, idempotencyMiddleware)
	}

	authenticated := apiV1.Group("")
	authenticated.Use(apiKeyMiddleware.Authenticate())
	authenticated.Use(authMiddleware.ValidateToken())
	authenticated.Use(middleware.CSRFMiddleware(logger))
	authenticated.Use(apiKeyMiddleware.RequireScopeAuto())
	authenticated.Use(apiKeyMiddleware.RequireUserSession())
	authenticated.Use(middleware.UserStatusMiddleware(userModule.UserRepo, logger))
	if authLimiter != nil {
		authenticated.Use(authLimiter)
	}
	{
		// Manually register auth routes that need authentication
		authGroup := authenticated.Group("/auth")
		authGroup.POST("/logout", authModule.AuthController.Logout)
		authGroup.POST("/ticket", idempotencyMiddleware, authModule.AuthController.GetTicket)
		authGroup.POST("/resend-verification", idempotencyMiddleware, authModule.AuthController.ResendVerification)
		authGroup.GET("/me", authModule.AuthController.Me)

		// Stats Routes
		statsGroup := authenticated.Group("/stats")
		{
			statsGroup.GET("/summary", delivery.AdaptHTTPHandler(statsModule.StatsController.HTTPSummary))
			statsGroup.GET("/activity", delivery.AdaptHTTPHandler(statsModule.StatsController.HTTPActivity))
			statsGroup.GET("/insights", delivery.AdaptHTTPHandler(statsModule.StatsController.HTTPInsights))
		}

		userHttp.RegisterAuthenticatedRoutes(authenticated, userModule.UserController)
		organizationHttp.RegisterAuthenticatedRoutes(authenticated, organizationModule.OrganizationController, idempotencyMiddleware)
		permissionHttp.RegisterBatchCheckRoute(authenticated, permissionModule.PermissionController)
		api_keyHttp.RegisterApiKeyRoutes(authenticated, apiKeyModule.Controller, authMiddleware, tenantMiddleware, idempotencyMiddleware)
	}

	tenantAuthorized := apiV1.Group("")
	tenantAuthorized.Use(apiKeyMiddleware.Authenticate())
	tenantAuthorized.Use(authMiddleware.ValidateToken())
	tenantAuthorized.Use(middleware.CSRFMiddleware(logger))
	tenantAuthorized.Use(apiKeyMiddleware.RequireScopeAuto())
	tenantAuthorized.Use(middleware.UserStatusMiddleware(userModule.UserRepo, logger))
	tenantAuthorized.Use(tenantMiddleware.RequireOrganization())
	tenantAuthorized.Use(casbinMiddleware)
	if authLimiter != nil {
		tenantAuthorized.Use(authLimiter)
	}
	{
		organizationHttp.RegisterTenantRoutes(tenantAuthorized, organizationModule.OrganizationController, apiKeyMiddleware, idempotencyMiddleware)
		roleHttp.RegisterTenantRoutes(tenantAuthorized, roleModule.RoleController, apiKeyMiddleware)

		// Project Routes
		projectGroup := tenantAuthorized.Group("/projects")
		{
			projectGroup.POST("", idempotencyMiddleware, apiKeyMiddleware.RequireScopes("project:manage"), delivery.AdaptHTTPHandler(projectModule.ProjectController.HTTPCreate))
			projectGroup.GET("", apiKeyMiddleware.RequireScopes("project:view", "project:manage"), delivery.AdaptHTTPHandler(projectModule.ProjectController.HTTPGetAll))
			projectGroup.GET("/:id", apiKeyMiddleware.RequireScopes("project:view", "project:manage"), delivery.AdaptHTTPHandler(projectModule.ProjectController.HTTPGetByID))
			projectGroup.PUT("/:id", idempotencyMiddleware, apiKeyMiddleware.RequireScopes("project:manage"), delivery.AdaptHTTPHandler(projectModule.ProjectController.HTTPUpdate))
			projectGroup.DELETE("/:id", apiKeyMiddleware.RequireScopes("project:manage"), delivery.AdaptHTTPHandler(projectModule.ProjectController.HTTPDelete))
		}

		webhookHttp.RegisterWebhookRoutes(tenantAuthorized, webhookModule.Controller, apiKeyMiddleware, idempotencyMiddleware)
	}

	authorized := apiV1.Group("")
	authorized.Use(apiKeyMiddleware.Authenticate())
	authorized.Use(authMiddleware.ValidateToken())
	authorized.Use(middleware.CSRFMiddleware(logger))
	authorized.Use(apiKeyMiddleware.RequireScopes("admin:manage"))
	authorized.Use(middleware.UserStatusMiddleware(userModule.UserRepo, logger))
	authorized.Use(tenantMiddleware.OptionalOrganization())
	authorized.Use(casbinMiddleware)
	if authLimiter != nil {
		authorized.Use(authLimiter)
	}
	{
		organizationHttp.RegisterAdminRoutes(authorized, organizationModule.OrganizationController, apiKeyMiddleware)
		permissionHttp.RegisterPermissionRoutes(authorized, permissionModule.PermissionController)
		accessHttp.RegisterAccessRoutes(authorized.Group("", tenantMiddleware.OptionalOrganization()), accessModule.AccessController, idempotencyMiddleware)
		roleHttp.RegisterAuthorizedRoutes(authorized, roleModule.RoleController)
		userHttp.RegisterAuthorizedRoutes(authorized, userModule.UserController)
		auditHttp.RegisterAuthorizedRoutes(authorized, auditModule.AuditController)
	}

	// TUS Upload Handler
	uploadGroup := router.Group("/api/v1/upload")
	uploadGroup.Use(authMiddleware.ValidateToken())
	uploadGroup.Use(middleware.UserStatusMiddleware(userModule.UserRepo, logger))
	{
		uploadGroup.Any("/files/*any", gin.WrapH(http.StripPrefix("/api/v1/upload/files/", tusHandler)))
	}

	return router
}

// GetHealth returns the health status of the application and its core dependencies.
func GetHealth(db any, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := "OK"
		details := make(map[string]string)

		if db != nil {
			if sqlxDB := tx.ExtractSQLX(db); sqlxDB != nil {
				if err := sqlxDB.PingContext(c.Request.Context()); err != nil {
					status = "DEGRADED"
					details["mysql"] = "DOWN"
				} else {
					details["mysql"] = "UP"
				}
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
	}
}
