package cardshttp

import (
	"student-kanban/internal/adapters/http_server/middleware"

	"github.com/gin-gonic/gin"
)

type CardController interface {
	CreateCard(c *gin.Context)
	GetCard(c *gin.Context)
	GetCardsByList(c *gin.Context)
	UpdateCard(c *gin.Context)
	DeleteCard(c *gin.Context)
	MoveCard(c *gin.Context)
}

// SetupCardRoutes registers all card-related routes
func SetupCardRoutes(r *gin.Engine, cardController CardController) {
	authorized := r.Group("/api/cards")
	authorized.Use(middleware.AuthMiddleware())
	{
		authorized.POST("", cardController.CreateCard)
		authorized.GET("/:id", cardController.GetCard)
		authorized.PATCH("/:id", cardController.UpdateCard)
		authorized.PATCH("/:id/position", cardController.MoveCard)
		authorized.DELETE("/:id", cardController.DeleteCard)
	}

	lists := r.Group("/api/lists")
	lists.Use(middleware.AuthMiddleware())
	{
		lists.GET("/:id/cards", cardController.GetCardsByList)
	}
}
