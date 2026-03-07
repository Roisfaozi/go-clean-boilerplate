package router

import (
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access"
	accessHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/delivery/http"
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
	_ "github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
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
	authMiddleware *middleware.AuthMiddleware,
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

	router.Use(gin.Recovery())
	router.Use(middleware.RequestIDMiddleware())

	if cfg.MetricsEnabled {
		router.Use(middleware.PrometheusMiddleware())
	}
	router.GET("/ws", authMiddleware.ValidateWebSocketToken(), wsController.HandleWebSocket)
	router.GET("/events", authMiddleware.ValidateToken(), sseManager.ServeHTTP())

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

	router.GET("/api/health", GetHealth(db, redisClient))

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
		authGroup.POST("/register", authModule.AuthController.Register)
		authGroup.GET("/sso/:provider", authModule.AuthController.SSOLogin)
		authGroup.GET("/sso/:provider/callback", authModule.AuthController.SSOCallback)

		userHttp.RegisterPublicRoutes(public, userModule.UserController)
		organizationHttp.RegisterPublicRoutes(public, organizationModule.OrganizationController)
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
		authGroup.POST("/ticket", authModule.AuthController.GetTicket)
		authGroup.POST("/resend-verification", authModule.AuthController.ResendVerification)
		authGroup.GET("/me", authModule.AuthController.Me)

		// Stats Routes
		statsGroup := authenticated.Group("/stats")
		{
			statsGroup.GET("/summary", statsModule.StatsController.GetSummary)
			statsGroup.GET("/activity", statsModule.StatsController.GetActivity)
			statsGroup.GET("/insights", statsModule.StatsController.GetInsights)
		}

		userHttp.RegisterAuthenticatedRoutes(authenticated, userModule.UserController)
		organizationHttp.RegisterAuthenticatedRoutes(authenticated, organizationModule.OrganizationController)
		permissionHttp.RegisterBatchCheckRoute(authenticated, permissionModule.PermissionController)
	}

	tenant := apiV1.Group("")
	tenant.Use(authMiddleware.ValidateToken())
	tenant.Use(tenantMiddleware.RequireOrganization())
	if authLimiter != nil {
		tenant.Use(authLimiter)
	}
	{
		organizationHttp.RegisterTenantRoutes(tenant, organizationModule.OrganizationController)

		// Project Routes
		projectGroup := tenant.Group("/projects")
		{
			projectGroup.POST("", projectModule.ProjectController.Create)
			projectGroup.GET("", projectModule.ProjectController.GetAll)
			projectGroup.GET("/:id", projectModule.ProjectController.GetByID)
			projectGroup.PUT("/:id", projectModule.ProjectController.Update)
			projectGroup.DELETE("/:id", projectModule.ProjectController.Delete)
		}
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

	// TUS Upload Handler
	uploadGroup := router.Group("/api/v1/upload")
	uploadGroup.Use(authMiddleware.ValidateToken())
	{
		uploadGroup.Any("/files/*any", gin.WrapH(http.StripPrefix("/api/v1/upload/files/", tusHandler)))
	}

	return router
}

// GetHealth godoc
// @Summary      System Health Check
// @Description  Returns the health status of the application and its core dependencies (MySQL, Redis).
// @Tags         system
// @Produce      json
// @Success      200  {object}  response.SwaggerGeneralResponseWrapper "System is healthy"
// @Failure      503  {object}  response.SwaggerErrorResponseWrapper "System is degraded"
// @Router       /health [get]
func GetHealth(db *gorm.DB, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
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
	}
}
