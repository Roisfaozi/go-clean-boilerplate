package config

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func NewServer(config *AppConfig, log *logrus.Logger) *gin.Engine {
	if config.Server.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()

	// Use Custom Structured Logger
	router.Use(middleware.RequestLogger(log))
	router.Use(gin.Recovery())

	// Basic CORS setup (can be refined later)
	router.Use(cors.Default())

	return router
}
