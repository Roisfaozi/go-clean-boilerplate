package config

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/audit"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role"
	roleRepository "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/router"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/jwt"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/tx"
	ws2 "github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/casbin/casbin/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Application holds all major application components.
type Application struct {
	Server   *http.Server
	DB       *gorm.DB
	Redis    *redis.Client
	Log      *logrus.Logger
	Enforcer *casbin.Enforcer
}

// NewApplication initializes and wires up all application components.
func NewApplication(cfg *AppConfig) (*Application, error) {
	logger := NewLogrus(cfg)
	validate := NewValidator()
	dbConnection := NewDatabase(cfg, logger)

	redisClient := NewRedisConfig(cfg, logger)

	tm := tx.NewTransactionManager(dbConnection, logger)

	jwtManager := jwt.NewJWTManager(
		cfg.JWT.AccessTokenSecret,
		cfg.JWT.RefreshTokenSecret,
		cfg.JWT.AccessTokenDuration,
		cfg.JWT.RefreshTokenDuration,
	)
	wsConfig := NewDefaultWebSocketConfig()
	wsManager := ws2.NewWebSocketManager((*ws2.WebSocketConfig)(wsConfig), logger, redisClient)
	wsController := ws2.NewWebSocketController(logger, wsManager, cfg.CORS.AllowedOrigins)
	go wsManager.Run()
	logger.Info("Shared dependencies initialized.")

	sseManager := sse.NewManager()
	logger.Info("SSE Manager initialized.")

	enforcer, err := NewCasbinEnforcer(cfg, dbConnection, logger)
	if err != nil {
		logger.Errorf("Error initializing casbin enforcer: %v", err)
		return nil, err
	}

	roleRepo := roleRepository.NewRoleRepository(dbConnection, logger)

	// Audit Module (Initialize early to inject into others)
	auditModule := audit.NewAuditModule(dbConnection, logger)

	authModule := auth.NewAuthModule(jwtManager, dbConnection, redisClient, logger, validate, tm, wsManager, enforcer, auditModule)

	userModule := user.NewUserModule(dbConnection, logger, validate, tm, enforcer, auditModule)

	permissionModule := permission.NewPermissionModule(enforcer, validate, logger, roleRepo)

	roleModule := role.NewRoleModule(dbConnection, logger, validate, tm)

	accessModule := access.NewAccessModule(dbConnection, logger, validate)

	logger.Info("Application modules initialized.")

	// Access AuthUseCase via AuthController
	authUseCase := authModule.AuthController.AuthUseCase
	authMiddleware := middleware.NewAuthMiddleware(authUseCase, logger)
	casbinMiddleware := middleware.CasbinMiddleware(enforcer, logger)
	logger.Info("Middleware initialized.")

	ginRouter := router.SetupRouter(
		router.RouterConfig{
			AllowedOrigins:   cfg.CORS.AllowedOrigins,
			TrustedProxies:   cfg.Server.TrustedProxies,
			RateLimitEnabled: cfg.RateLimit.Enabled,
			RateLimitRPS:     cfg.RateLimit.RPS,
			RateLimitBurst:   cfg.RateLimit.Burst,
			RateLimitStore:   cfg.RateLimit.Store,
		},
		authModule,
		userModule,
		permissionModule,
		accessModule,
		roleModule,
		auditModule,
		authMiddleware,
		casbinMiddleware,
		wsController,
		sseManager,
		redisClient,
		logger,
	)
	logger.Info("Router setup complete.")

	serverPort := fmt.Sprintf(":%d", cfg.Server.Port)
	httpServer := &http.Server{
		Addr:    serverPort,
		Handler: ginRouter,
	}
	logger.Infof("Server configured to run on port %s", serverPort)

	app := &Application{
		Server:   httpServer,
		DB:       dbConnection,
		Redis:    redisClient,
		Log:      logger,
		Enforcer: enforcer,
	}

	return app, nil
}

// Shutdown gracefully shuts down all application components.
func (app *Application) Shutdown(ctx context.Context) error {
	app.Log.Info("Shutting down HTTP server...")
	if err := app.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	if app.Redis != nil {
		app.Log.Info("Closing Redis connection...")
		if err := app.Redis.Close(); err != nil {
			app.Log.Errorf("Failed to close Redis client: %v", err)
		}
	}

	if app.DB != nil {
		app.Log.Info("Closing database connection...")
		sqlDB, err := app.DB.DB()
		if err != nil {
			app.Log.Errorf("Failed to get DB instance for closing: %v", err)
		} else if err := sqlDB.Close(); err != nil {
			app.Log.Errorf("Failed to close database connection: %v", err)
		}
	}

	return nil
}