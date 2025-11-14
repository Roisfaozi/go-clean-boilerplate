package router

import (
	"github.com/Roisfaozi/casbin-db/internal/middleware"
	"github.com/Roisfaozi/casbin-db/internal/modules/auth"
	authHttp "github.com/Roisfaozi/casbin-db/internal/modules/auth/delivery/http"
	"github.com/Roisfaozi/casbin-db/internal/modules/user"
	userHttp "github.com/Roisfaozi/casbin-db/internal/modules/user/delivery/http"
	"github.com/Roisfaozi/casbin-db/internal/utils/ws"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/Roisfaozi/casbin-db/docs"
)

// SetupRouter initializes the Gin router and registers all application routes.
func SetupRouter(
	authModule *auth.AuthModule,
	userModule *user.UserModule,
	authMiddleware *middleware.AuthMiddleware,
	wsController *ws.WebSocketController,
) *gin.Engine {
	router := gin.New()

	// Global Middlewares
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	// API v1 Group
	apiV1 := router.Group("/api/v1")

	// Register module routes
	authHttp.RegisterAuthRoutes(apiV1, authModule.AuthHandler(), authMiddleware)
	userHttp.RegisterUserRoutes(apiV1, userModule.UserHandler(), authMiddleware)

	// Register WebSocket route
	router.GET("/ws", wsController.HandleWebSocket)

	// Register Swagger route
	router.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
