package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Roisfaozi/casbin-db/internal/config"
)

// @title           Modular API with JWT
// @version         1.0
// @description     This is a sample server for a modular application with JWT authentication.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description "Type 'Bearer ' followed by a space and the access token."
func main() {
	// 1. Initialize Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	if cfg.JWT.AccessTokenSecret == "" || cfg.JWT.RefreshTokenSecret == "" {
		log.Fatal("JWT secrets are not set. Please check your .env file or environment variables.")
	}

	// 2. Create the application
	app, err := config.NewApplication(cfg)
	if err != nil {
		log.Fatalf("Failed to create application: %v", err)
	}

	// 3. Create a context that is canceled on a SIGINT or SIGTERM signal
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 4. Start the server in a goroutine
	go func() {
		log.Printf("Starting server on %s", app.Server.Addr)
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// 5. Wait for the shutdown signal
	<-ctx.Done()
	log.Println("Shutting down server...")

	// 6. Create a context with a timeout for the shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server exiting")
}
