package boardhttp

import (
	"student-kanban/internal/adapters/http_server/middleware"

	"github.com/gin-gonic/gin"
)

type BoardController interface {
	CreateBoard(ctx *gin.Context)
	GetBoard(ctx *gin.Context)
	GetUserBoards(ctx *gin.Context)
	UpdateBoard(ctx *gin.Context)
	DeleteBoard(ctx *gin.Context)
}

func SetupBoardRoutes(r *gin.Engine, boardController BoardController) {
	authorized := r.Group("/api/boards")
	authorized.Use(middleware.AuthMiddleware())
	{
		authorized.POST("", boardController.CreateBoard)
		authorized.GET("", boardController.GetUserBoards)
		authorized.GET("/:id", boardController.GetBoard)
		authorized.PATCH("/:id", boardController.UpdateBoard)
		authorized.DELETE("/:id", boardController.DeleteBoard)
	}

}
