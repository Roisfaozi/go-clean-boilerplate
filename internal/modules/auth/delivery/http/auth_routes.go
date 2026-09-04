package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/delivery"
	"github.com/gin-gonic/gin"
)

func RegisterPublicRoutes(router *gin.RouterGroup, controller *AuthController) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/register", delivery.AdaptHTTPHandler(controller.HTTPRegister))
		authGroup.POST("/login", delivery.AdaptHTTPHandler(controller.HTTPLogin))
		authGroup.POST("/refresh", delivery.AdaptHTTPHandler(controller.HTTPRefreshToken))
		authGroup.POST("/forgot-password", delivery.AdaptHTTPHandler(controller.HTTPForgotPassword))
		authGroup.POST("/reset-password", delivery.AdaptHTTPHandler(controller.HTTPResetPassword))
		authGroup.POST("/verify-email", delivery.AdaptHTTPHandler(controller.HTTPVerifyEmail))

		// SSO Routes
		authGroup.GET("/sso/:provider", delivery.AdaptHTTPHandler(controller.HTTPSSOLogin))
		authGroup.GET("/sso/:provider/callback", delivery.AdaptHTTPHandler(controller.HTTPSSOCallback))
	}
}

func RegisterAuthenticatedRoutes(router *gin.RouterGroup, controller *AuthController) {
	authGroup := router.Group("/auth")
	{
		authGroup.POST("/logout", delivery.AdaptHTTPHandler(controller.HTTPLogout))
		authGroup.POST("/resend-verification", delivery.AdaptHTTPHandler(controller.HTTPResendVerification))
		authGroup.GET("/me", delivery.AdaptHTTPHandler(controller.HTTPMe))
		authGroup.POST("/ticket", delivery.AdaptHTTPHandler(controller.HTTPGetTicket))
	}
}
