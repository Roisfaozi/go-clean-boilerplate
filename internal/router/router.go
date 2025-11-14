package router

import (
	"github.com/Roisfaozi/casbin-db/internal/middleware"
	"github.com/Roisfaozi/casbin-db/internal/modules/auth"
	authHttp "github.com/Roisfaozi/casbin-db/internal/modules/auth/delivery/http"
	"github.com/Roisfaozi/casbin-db/internal/modules/user"
	userHttp "github.com/Roisfaozi/casbin-db/internal/modules/user/delivery/http"
	"github.com/Roisfaozi/casbin-db/internal/utils/ws"
	"github.com/gin-gonic/gin"
)

func SetupRouter(
	authModule *auth.AuthModule,
	userModule *user.UserModule,
	authMiddleware *middleware.AuthMiddleware,
	wsController *ws.WebSocketController,
) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	apiV1 := router.Group("/api/v1")

	authHttp.RegisterAuthRoutes(apiV1, authModule.AuthHandler(), authMiddleware)
	userHttp.RegisterUserRoutes(apiV1, userModule.UserHandler(), authMiddleware)

	router.GET("/ws", wsController.HandleWebSocket)

	return router
}
