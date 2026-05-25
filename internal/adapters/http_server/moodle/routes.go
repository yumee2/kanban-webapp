package moodlehttp

import (
	"student-kanban/internal/adapters/http_server/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, controller *Controller) {
	authorized := r.Group("/api/integrations/moodle")
	authorized.Use(middleware.AuthMiddleware())
	{
		authorized.POST("/connect", controller.Connect)
		authorized.GET("/connection", controller.GetConnection)
		authorized.GET("/courses", controller.ListCourses)
		authorized.POST("/import-board", controller.ImportBoard)
	}
}
