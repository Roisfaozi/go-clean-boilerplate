package http

import (
	"github.com/gin-gonic/gin"
)

func RegisterAuthorizedRoutes(router *gin.RouterGroup, roleHandler *RoleController) {
	roleGroup := router.Group("/roles")
	{
		roleGroup.POST("", roleHandler.Create)
		roleGroup.GET("", roleHandler.GetAll)
		roleGroup.PUT("/:id", roleHandler.Update)
		roleGroup.POST("/search", roleHandler.GetRolesDynamic)
		roleGroup.DELETE("/:id", roleHandler.Delete)
	}
}
