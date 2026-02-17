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
func (h *ListHandler) SetupListRoutes(r *gin.Engine, listController ListController) {
	authorized := r.Group("/api/lists")
	authorized.Use(middleware.AuthMiddleware())
	{
		authorized.POST("", h.CreateList)
		authorized.GET("/:id", h.GetList)
		authorized.PUT("/:id", h.UpdateList)
		authorized.DELETE("/:id", h.DeleteList)
		authorized.PATCH("/:id/position", h.UpdatePosition)
	}

	// Nested route under boards
	boards := r.Group("/api/boards")
	boards.Use(middleware.AuthMiddleware())
	{
		boards.GET("/:id/lists", h.GetListsByBoard)
	}
}
