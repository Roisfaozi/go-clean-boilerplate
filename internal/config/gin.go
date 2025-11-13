package config

import "github.com/gin-gonic/gin"

func NewServer(config *AppConfig) *gin.Engine {
	if config.Server.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	return router
}
