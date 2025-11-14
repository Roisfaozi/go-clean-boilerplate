package config

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Roisfaozi/casbin-db/internal/middleware"
	"github.com/Roisfaozi/casbin-db/internal/modules/auth"
	"github.com/Roisfaozi/casbin-db/internal/modules/user"
	"github.com/Roisfaozi/casbin-db/internal/router"
	"github.com/Roisfaozi/casbin-db/internal/utils/jwt"
	"github.com/Roisfaozi/casbin-db/internal/utils/tx"
	"github.com/Roisfaozi/casbin-db/internal/utils/ws"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Application holds all major application components.
type Application struct {
	Server *http.Server
	DB     *gorm.DB
	Redis  *redis.Client
	Log    *logrus.Logger
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

	// 2. Initialize Modules
	authModule := auth.NewAuthModule(jwtManager, dbConnection, redisClient, logger, validate, tm, wsManager)
	userModule := user.NewUserModule(dbConnection, logger, validate, tm)
	logger.Info("Application modules initialized.")

	// 3. Initialize Middleware
	authUseCase := authModule.AuthHandler().AuthUseCase
	authMiddleware := middleware.NewAuthMiddleware(authUseCase, logger)
	logger.Info("Middleware initialized.")

	// 4. Setup Router
	ginRouter := router.SetupRouter(authModule, userModule, authMiddleware, wsController)
	logger.Info("Router setup complete.")

	// 5. Create HTTP Server
	serverPort := fmt.Sprintf(":%d", cfg.Server.Port)
	httpServer := &http.Server{
		Addr:    serverPort,
		Handler: ginRouter,
	}
	logger.Infof("Server configured to run on port %s", serverPort)

	// 6. Return the application container
	app := &Application{
		Server: httpServer,
		DB:     dbConnection,
		Redis:  redisClient,
		Log:    logger,
	}

	return app, nil
}

// Shutdown gracefully shuts down all application components.
func (app *Application) Shutdown(ctx context.Context) error {
	app.Log.Info("Shutting down HTTP server...")
	if err := app.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server forced to shutdown: %w", err)
	}

	app.Log.Info("Closing Redis connection...")
	if err := app.Redis.Close(); err != nil {
		return fmt.Errorf("failed to close Redis client: %w", err)
	}

	app.Log.Info("Closing database connection...")
	sqlDB, err := app.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get DB instance for closing: %w", err)
	}
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}

	return nil
}
