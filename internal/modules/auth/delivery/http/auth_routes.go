package http

import (
	"github.com/gin-gonic/gin"
)

func RegisterPublicRoutes(router *gin.RouterGroup, controller *AuthController) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", controller.Register)
		authGroup.POST("/login", controller.Login)
		authGroup.POST("/refresh", controller.RefreshToken)
		authGroup.POST("/forgot-password", controller.ForgotPassword)
		authGroup.POST("/reset-password", controller.ResetPassword)
		authGroup.POST("/verify-email", controller.VerifyEmail)

		// SSO Routes
		authGroup.GET("/sso/:provider", controller.SSOLogin)
		authGroup.GET("/sso/:provider/callback", controller.SSOCallback)
	}
}

func RegisterAuthenticatedRoutes(router *gin.RouterGroup, controller *AuthController) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/logout", controller.Logout)
		authGroup.POST("/resend-verification", controller.ResendVerification)
	}
}
