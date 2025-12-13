package router

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/middleware"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access"
	accessHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/access/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth"
	authHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/auth/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission"
	permissionHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/permission/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role"
	roleHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/role/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user"
	userHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/user/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/sse"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/ws"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter(
	authModule *auth.AuthModule,
	userModule *user.UserModule,
	permissionModule *permission.PermissionModule,
	accessModule *access.AccessModule,
	roleModule *role.RoleModule,
	authMiddleware *middleware.AuthMiddleware,
	casbinMiddleware gin.HandlerFunc,
	wsController *ws.WebSocketController,
	sseManager *sse.Manager,
	logger *logrus.Logger,
) *gin.Engine {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.RecoveryMiddleware(logger))
	router.Use(middleware.SecurityMiddleware())
	router.Use(middleware.CORSMiddleware())

	router.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
		})
	})

	router.GET("/ws", wsController.HandleWebSocket)
	router.GET("/events", sseManager.ServeHTTP())

	apiV1 := router.Group("/api/v1")

	public := apiV1.Group("")
	{
		authHttp.RegisterPublicRoutes(public, authModule.AuthHandler())
		userHttp.RegisterPublicRoutes(public, userModule.UserHandler())
	}

	authenticated := apiV1.Group("")
	authenticated.Use(authMiddleware.ValidateToken())
	{
		authHttp.RegisterAuthenticatedRoutes(authenticated, authModule.AuthHandler())
	}

	authorized := apiV1.Group("")
	authorized.Use(authMiddleware.ValidateToken())
	authorized.Use(casbinMiddleware)
	{
		userHttp.RegisterAuthorizedRoutes(authorized, userModule.UserHandler())
		permissionHttp.RegisterPermissionRoutes(authorized, permissionModule.PermissionHandler())
		accessHttp.RegisterAccessRoutes(authorized, accessModule.AccessHandler())
		roleHttp.RegisterAuthorizedRoutes(authorized, roleModule.RoleHandler())
	}

	return router
}
