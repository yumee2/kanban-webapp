package listshttp

import (
	"student-kanban/internal/adapters/http_server/middleware"

	"github.com/gin-gonic/gin"
)

type ListController interface {
	GetList(c *gin.Context)
	CreateList(c *gin.Context)
	GetListsByBoard(c *gin.Context)
	UpdateList(c *gin.Context)
	DeleteList(c *gin.Context)
	UpdatePosition(c *gin.Context)
}

// RegisterRoutes registers all list-related routes
func SetupListRoutes(r *gin.Engine, listController ListController) {
	authorized := r.Group("/api/lists")
	authorized.Use(middleware.AuthMiddleware())
	{
		authorized.POST("", listController.CreateList)
		authorized.GET("/:id", listController.GetList)
		authorized.PUT("/:id", listController.UpdateList)
		authorized.DELETE("/:id", listController.DeleteList)
		authorized.PATCH("/:id/position", listController.UpdatePosition)
	}

	// Nested route under boards
	boards := r.Group("/api/boards")
	boards.Use(middleware.AuthMiddleware())
	{
		boards.GET("/:id/lists", listController.GetListsByBoard)
	}
}
