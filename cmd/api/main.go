package main

import (
	"log"

	"github.com/Roisfaozi/casbin-db/internal/config"
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

	// 2. Create and run the server
	server, err := NewServer(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
