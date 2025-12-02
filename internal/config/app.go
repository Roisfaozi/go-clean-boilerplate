package config

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Roisfaozi/casbin-db/internal/middleware"
	"github.com/Roisfaozi/casbin-db/internal/modules/access"
	"github.com/Roisfaozi/casbin-db/internal/modules/auth"
	"github.com/Roisfaozi/casbin-db/internal/modules/permission"
	"github.com/Roisfaozi/casbin-db/internal/modules/role"
	roleRepository "github.com/Roisfaozi/casbin-db/internal/modules/role/repository"
	"github.com/Roisfaozi/casbin-db/internal/modules/user"
	"github.com/Roisfaozi/casbin-db/internal/router"
	"github.com/Roisfaozi/casbin-db/internal/utils/jwt"
	"github.com/Roisfaozi/casbin-db/internal/utils/tx"
	"github.com/Roisfaozi/casbin-db/internal/utils/ws"
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
	// 1. Initialize Shared Dependencies
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
	wsManager := ws.NewWebSocketManager((*ws.WebSocketConfig)(wsConfig), logger)
	wsController := ws.NewWebSocketController(logger, wsManager)
	go wsManager.Run()
	logger.Info("Shared dependencies initialized.")

	// 2. Initialize Casbin (conditionally)
	enforcer, err := NewCasbinEnforcer(cfg, dbConnection, logger)
	if err != nil {
		return nil, err
	}

	// 3. Initialize Modules
	roleRepo := roleRepository.NewRoleRepository(dbConnection, logger)

	authModule := auth.NewAuthModule(jwtManager, dbConnection, redisClient, logger, validate, tm, wsManager, enforcer)

	userModule := user.NewUserModule(dbConnection, logger, validate, tm, enforcer)

	permissionModule := permission.NewPermissionModule(enforcer, validate, logger, roleRepo)

	roleModule := role.NewRoleModule(dbConnection, logger, validate, tm)

	accessModule := access.NewAccessModule(dbConnection, logger, validate)

	logger.Info("Application modules initialized.")

	// 4. Initialize Middleware
	authUseCase := authModule.AuthHandler().AuthUseCase
	authMiddleware := middleware.NewAuthMiddleware(authUseCase, logger)
	casbinMiddleware := middleware.CasbinMiddleware(enforcer, logger)
	logger.Info("Middleware initialized.")

	// 5. Setup Router
	ginRouter := router.SetupRouter(
		authModule,
		userModule,
		permissionModule,
		accessModule,
		roleModule,
		authMiddleware,
		casbinMiddleware,
		wsController,
	)
	logger.Info("Router setup complete.")

	// 6. Create HTTP Server
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
