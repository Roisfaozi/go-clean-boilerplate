package main

import (
	"fmt"

	"github.com/Roisfaozi/casbin-db/internal/config"
	authHttp "github.com/Roisfaozi/casbin-db/internal/modules/auth/delivery/http"
	"github.com/Roisfaozi/casbin-db/internal/modules/auth/delivery/http/middleware"
	authRepo "github.com/Roisfaozi/casbin-db/internal/modules/auth/repository"
	authUsecase "github.com/Roisfaozi/casbin-db/internal/modules/auth/usecase"
	userRepo "github.com/Roisfaozi/casbin-db/internal/modules/user/repository"
	"github.com/Roisfaozi/casbin-db/internal/utils/tx"
	"github.com/Roisfaozi/casbin-db/internal/utils/ws"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Server holds all dependencies for this application.
type Server struct {
	router    *gin.Engine
	log       *logrus.Logger
	cfg       *config.AppConfig
	db        *gorm.DB
	wsManager ws.Manager
}

// NewServer creates a new server instance with all dependencies initialized.
func NewServer(cfg *config.AppConfig) (*Server, error) {
	// Initialize Logger
	logger := config.NewLogrus(cfg)
	logger.Info("Configuration and logger initialized successfully.")

	// Initialize Validator
	validate := config.NewValidator()
	logger.Info("Validator initialized.")

	// Initialize Database Connections
	dbConnection := config.NewDatabase(cfg, logger)

	redisClient := config.NewRedisConfig(cfg, logger)

	logger.Info("Database connections established.")

	// Initialize WebSocket Manager
	wsConfig := config.NewDefaultWebSocketConfig()
	wsManager := ws.NewWebSocketManager((*ws.WebSocketConfig)(wsConfig), logger)
	wsController := ws.NewWebSocketController(logger, wsManager)
	logger.Info("WebSocket manager and controller initialized.")

	// Initialize Transaction Manager
	tm := tx.NewTransactionManager(dbConnection, logger)
	logger.Info("Transaction manager initialized.")

	// Initialize Repositories
	tokenRepository := authRepo.NewTokenRepositoryRedis(redisClient, logger)
	userRepository := userRepo.NewUserRepository(dbConnection, logger)
	logger.Info("Repositories initialized.")

	// Initialize Use Cases (Services)
	authService := authUsecase.NewService(cfg, tokenRepository, userRepository, validate, tm, logger, wsManager)
	logger.Info("Use cases initialized.")

	// Initialize Handlers and Middleware
	authAPIHandler := authHttp.NewAuthHandler(authService, logger)
	authMiddleware := middleware.NewAuthMiddleware(authService, logger)
	logger.Info("Handlers and middleware initialized.")

	// Setup Router and Routes
	router := gin.Default()
	api := router.Group("/api/v1")
	authHttp.RegisterAuthRoutes(api, authAPIHandler, authMiddleware)
	logger.Info("Auth routes registered.")

	// Register WebSocket route
	router.GET("/ws", wsController.HandleWebSocket)
	logger.Info("WebSocket route registered at /ws.")

	server := &Server{
		router:    router,
		log:       logger,
		cfg:       cfg,
		db:        dbConnection,
		wsManager: wsManager,
	}

	return server, nil
}

// Start runs the HTTP server and WebSocket manager.
func (s *Server) Start() error {
	// Run WebSocket manager in a separate goroutine
	go s.wsManager.Run()

	serverPort := fmt.Sprintf(":%d", s.cfg.Server.Port)
	s.log.Infof("Starting server on port %s", serverPort)
	return s.router.Run(serverPort)
}
