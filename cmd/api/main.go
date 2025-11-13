package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Roisfaozi/casbin-db/internal/config"
	// Import placeholder for your future packages
	// authHandler "github.com/Roisfaozi/casbin-db/internal/modules/auth/handler"
	// authRepo "github.com/Roisfaozi/casbin-db/internal/modules/auth/repository"
	// authUsecase "github.com/Roisfaozi/casbin-db/internal/modules/auth/usecase"
	// userRepo "github.com/Roisfaozi/casbin-db/internal/modules/user/repository"
	// "github.com/Roisfaozi/casbin-db/internal/utils/db"
	// "github.com/Roisfaozi/casbin-db/internal/utils/tx"
)

func main() {
	// 1. Initialize Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	if cfg.JWT.AccessTokenSecret == "" || cfg.JWT.RefreshTokenSecret == "" {
		log.Fatal("JWT secrets are not set. Please check your .env file or environment variables.")
	}

	// 2. Initialize Logger
	logger := config.NewLogrus(cfg)
	logger.Info("Configuration and logger initialized successfully.")

	// 3. Initialize Validator
	//validate := validator.New()
	logger.Info("Validator initialized.")

	// --- Dependency Injection Wiring ---
	// The following sections are placeholders to show how you would wire your application.
	// You will need to create these packages and functions.

	// 4. Initialize Database Connections (e.g., PostgreSQL & Redis)
	// dbConnection, err := db.NewGormConnection(cfg)
	// if err != nil {
	// 	logger.Fatalf("Failed to connect to database: %v", err)
	// }
	// redisClient, err := db.NewRedisClient(cfg)
	// if err != nil {
	// 	logger.Fatalf("Failed to connect to Redis: %v", err)
	// }
	// logger.Info("Database connections established.")

	// 5. Initialize Transaction Manager
	// tm := tx.NewTransactionManager(dbConnection)
	// logger.Info("Transaction manager initialized.")

	// 6. Initialize Repositories
	// tokenRepository := authRepo.NewTokenRepository(redisClient)
	// userRepository := userRepo.NewUserRepository(dbConnection)
	// logger.Info("Repositories initialized.")

	// 7. Initialize Use Cases (Services)
	// Note: We can pass 'cfg' directly because it implements the required interface.
	// authService := authUsecase.NewService(cfg, tokenRepository, userRepository, validate, tm, logger)
	// logger.Info("Use cases initialized.")

	// 8. Initialize Handlers (Controllers)
	// authAPIHandler := authHandler.NewAuthHandler(authService, logger)
	// logger.Info("Handlers initialized.")

	// 9. Setup Router
	// router := http.NewServeMux()
	// authAPIHandler.RegisterRoutes(router)
	// For now, a simple root handler:
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "Server is running with new configuration!")
	})

	// 10. Start Server
	serverPort := fmt.Sprintf(":%d", cfg.Server.Port)
	logger.Infof("Starting server on port %s", serverPort)

	err = http.ListenAndServe(serverPort, nil)
	if err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
