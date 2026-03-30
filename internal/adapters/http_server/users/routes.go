package userhttp

import "github.com/gin-gonic/gin"

type UserController interface {
	Register(ctx *gin.Context)
	Login(ctx *gin.Context)
	Refresh(ctx *gin.Context)
	Logout(ctx *gin.Context)
}

func SetupAuthRoutes(r *gin.Engine, authController UserController) {
	urlGroup := r.Group("/api/auth")
	{
		urlGroup.POST("/register", authController.Register)
		urlGroup.POST("/login", authController.Login)
		urlGroup.POST("/refresh", authController.Refresh)
		urlGroup.POST("/logout", authController.Logout)
	}
}
