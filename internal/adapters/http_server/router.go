package httpserver

import "github.com/gin-gonic/gin"

type AuthController interface {
	Register(ctx *gin.Context)
	Login(ctx *gin.Context)
	Refresh(ctx *gin.Context)
}

func SetupAuthRoutes(r *gin.Engine, authController AuthController) {
	urlGroup := r.Group("/auth")
	{
		urlGroup.POST("/register", authController.Register)
		urlGroup.POST("/login", authController.Login)
		urlGroup.POST("/refresh", authController.Refresh)
	}
}
