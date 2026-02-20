package tagshttp

import (
	"student-kanban/internal/adapters/http_server/middleware"

	"github.com/gin-gonic/gin"
)

type TagController interface {
	CreateTag(c *gin.Context)
	GetTag(c *gin.Context)
	GetTagsByBoard(c *gin.Context)
	UpdateTag(c *gin.Context)
	DeleteTag(c *gin.Context)
	AttachTag(c *gin.Context)
	DetachTag(c *gin.Context)
	GetCardTags(c *gin.Context)
}

func SetupTagRoutes(r *gin.Engine, tagController TagController) {
	authorized := r.Group("/api/tags")
	authorized.Use(middleware.AuthMiddleware())
	{
		authorized.POST("", tagController.CreateTag)
		authorized.GET("/:id", tagController.GetTag)
		authorized.PATCH("/:id", tagController.UpdateTag)
		authorized.DELETE("/:id", tagController.DeleteTag)
	}

	boards := r.Group("/api/boards")
	boards.Use(middleware.AuthMiddleware())
	{
		boards.GET("/:id/tags", tagController.GetTagsByBoard)
	}

	cards := r.Group("/api/cards")
	cards.Use(middleware.AuthMiddleware())
	{
		cards.GET("/:id/tags", tagController.GetCardTags)
		cards.POST("/:id/tags/:tag_id", tagController.AttachTag)
		cards.DELETE("/:id/tags/:tag_id", tagController.DetachTag)
	}
}
