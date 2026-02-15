package boardhttp

import (
	"errors"
	"log/slog"
	"net/http"
	"student-kanban/internal/domain/entity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const UserIDKey = "userID"

type BoardService interface {
	CreateBoard(ownerID uuid.UUID, title, description string) (string, error)
	GetBoardsByOwnerID(ownerID uuid.UUID) ([]*entity.Board, error)
	GetBoardByID(id uuid.UUID) (*entity.Board, error)
	UpdateBoard(id uuid.UUID, title *string, description *string) error
	DeleteBoard(id uuid.UUID) error
}

type boardController struct {
	boardService BoardService
}

func NewBoardController(boardService BoardService) *boardController {
	return &boardController{
		boardService: boardService,
	}
}

// CreateBoard creates a new board for the authenticated user
func (c *boardController) CreateBoard(ctx *gin.Context) {
	const fn = "adapters.controller.CreateBoard"
	log := slog.With(slog.String("fn", fn))

	// Get user ID from context (set by auth middleware)
	userIDValue, exists := ctx.Get(UserIDKey)
	if !exists {
		log.Error("user ID not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		log.Error("invalid user ID type in context")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	var request createBoardRequest
	if err := ctx.BindJSON(&request); err != nil {
		log.Error("failed to parse json body", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	boardID, err := c.boardService.CreateBoard(userID, request.Title, request.Description)
	if err != nil {
		log.Error("failed to create board", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create board"})
		return
	}

	ctx.JSON(http.StatusCreated, createBoardResponse{
		ID: boardID,
	})
}

// GetBoard retrieves a board by ID
func (c *boardController) GetBoard(ctx *gin.Context) {
	const fn = "adapters.controller.GetBoard"
	log := slog.With(slog.String("fn", fn))

	boardIDParam := ctx.Param("id")
	boardID, err := uuid.Parse(boardIDParam)
	if err != nil {
		log.Error("invalid board ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	board, err := c.boardService.GetBoardByID(boardID)
	if err != nil {
		if errors.Is(err, entity.ErrBoardNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
			return
		}
		log.Error("failed to get board", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve board"})
		return
	}

	ctx.JSON(http.StatusOK, toBoardResponse(board))
}

// GetUserBoards retrieves all boards for the authenticated user
func (c *boardController) GetUserBoards(ctx *gin.Context) {
	const fn = "adapters.controller.GetUserBoards"
	log := slog.With(slog.String("fn", fn))

	// Get user ID from context
	userIDValue, exists := ctx.Get(UserIDKey)
	if !exists {
		log.Error("user ID not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		log.Error("invalid user ID type in context")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	boards, err := c.boardService.GetBoardsByOwnerID(userID)
	if err != nil {
		log.Error("failed to get user boards", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve boards"})
		return
	}

	response := make([]boardResponse, len(boards))
	for i, board := range boards {
		response[i] = toBoardResponse(board)
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateBoard updates a board's title and/or description
func (c *boardController) UpdateBoard(ctx *gin.Context) {
	const fn = "adapters.controller.UpdateBoard"
	log := slog.With(slog.String("fn", fn))

	boardIDParam := ctx.Param("id")
	boardID, err := uuid.Parse(boardIDParam)
	if err != nil {
		log.Error("invalid board ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	var request updateBoardRequest
	if err := ctx.BindJSON(&request); err != nil {
		log.Error("failed to parse json body", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate that at least one field is being updated
	if request.Title == nil && request.Description == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "At least one field must be provided"})
		return
	}

	err = c.boardService.UpdateBoard(boardID, request.Title, request.Description)
	if err != nil {
		if errors.Is(err, entity.ErrBoardNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
			return
		}
		log.Error("failed to update board", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update board"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Board updated successfully"})
}

// DeleteBoard deletes a board by ID
func (c *boardController) DeleteBoard(ctx *gin.Context) {
	const fn = "adapters.controller.DeleteBoard"
	log := slog.With(slog.String("fn", fn))

	boardIDParam := ctx.Param("id")
	boardID, err := uuid.Parse(boardIDParam)
	if err != nil {
		log.Error("invalid board ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	err = c.boardService.DeleteBoard(boardID)
	if err != nil {
		if errors.Is(err, entity.ErrBoardNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
			return
		}
		log.Error("failed to delete board", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete board"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Board deleted successfully"})
}

// Helper function to convert entity to response DTO
func toBoardResponse(board *entity.Board) boardResponse {
	return boardResponse{
		ID:          board.ID.String(),
		OwnerID:     board.OwnerID.String(),
		Title:       board.Title,
		Description: board.Description,
	}
}
