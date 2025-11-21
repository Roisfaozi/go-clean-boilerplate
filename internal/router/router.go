package router

import (
	"github.com/Roisfaozi/casbin-db/internal/middleware"
	"github.com/Roisfaozi/casbin-db/internal/modules/access"
	accessHttp "github.com/Roisfaozi/casbin-db/internal/modules/access/delivery/http"
	"github.com/Roisfaozi/casbin-db/internal/modules/auth"
	authHttp "github.com/Roisfaozi/casbin-db/internal/modules/auth/delivery/http"
	"github.com/Roisfaozi/casbin-db/internal/modules/permission"
	permissionHttp "github.com/Roisfaozi/casbin-db/internal/modules/permission/delivery/http"
	"github.com/Roisfaozi/casbin-db/internal/modules/role"
	"github.com/Roisfaozi/casbin-db/internal/modules/user"
	userHttp "github.com/Roisfaozi/casbin-db/internal/modules/user/delivery/http"
	"github.com/Roisfaozi/casbin-db/internal/utils/ws"
	"github.com/gin-gonic/gin"
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
) *gin.Engine {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.CORSMiddleware())

	router.GET("/api/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	router.GET("/ws", wsController.HandleWebSocket)

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
	}

	return router
}
