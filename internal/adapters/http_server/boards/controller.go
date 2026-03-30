package boardhttp

import (
	"errors"
	"log/slog"
	"net/http"
	"student-kanban/internal/domain/entity"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const UserIDKey = "userID"

type BoardService interface {
	CreateBoard(ownerID uuid.UUID, title, description string) (string, error)
	GetBoardsByOwnerID(ownerID uuid.UUID) ([]*entity.Board, error)
	GetBoardByIDForOwner(id uuid.UUID, ownerID uuid.UUID) (*entity.Board, error)
	UpdateBoard(ownerID uuid.UUID, id uuid.UUID, title *string, description *string) error
	DeleteBoard(ownerID uuid.UUID, id uuid.UUID) error
}

type boardController struct {
	boardService BoardService
}

func NewBoardController(boardService BoardService) *boardController {
	return &boardController{
		boardService: boardService,
	}
}

// CreateBoard godoc
// @Summary      Create a new board
// @Description  Creates a new board for the authenticated user
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        body  body      createBoardRequest   true  "Board data"
// @Success      201   {object}  createBoardResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /boards [post]
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

// GetBoard godoc
// @Summary      Get a board by ID
// @Description  Returns a single board by its UUID
// @Tags         boards
// @Produce      json
// @Param        id   path      string  true  "Board UUID"
// @Success      200  {object}  boardResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /boards/{id} [get]
func (c *boardController) GetBoard(ctx *gin.Context) {
	const fn = "adapters.controller.GetBoard"
	log := slog.With(slog.String("fn", fn))

	userID, err := getAuthorizedUserID(ctx, log)
	if err != nil {
		return
	}

	boardIDParam := ctx.Param("id")
	boardID, err := uuid.Parse(boardIDParam)
	if err != nil {
		log.Error("invalid board ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	board, err := c.boardService.GetBoardByIDForOwner(boardID, userID)
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

// GetUserBoards godoc
// @Summary      Get all boards for the authenticated user
// @Description  Returns a list of all boards owned by the authenticated user
// @Tags         boards
// @Produce      json
// @Success      200  {array}   boardResponse
// @Failure      401  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /boards [get]
func (c *boardController) GetUserBoards(ctx *gin.Context) {
	const fn = "adapters.controller.GetUserBoards"
	log := slog.With(slog.String("fn", fn))

	userID, err := getAuthorizedUserID(ctx, log)
	if err != nil {
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

// UpdateBoard godoc
// @Summary      Update a board
// @Description  Updates a board's title and/or description. At least one field must be provided
// @Tags         boards
// @Accept       json
// @Produce      json
// @Param        id    path      string             true  "Board UUID"
// @Param        body  body      updateBoardRequest  true  "Fields to update"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /boards/{id} [put]
func (c *boardController) UpdateBoard(ctx *gin.Context) {
	const fn = "adapters.controller.UpdateBoard"
	log := slog.With(slog.String("fn", fn))

	userID, err := getAuthorizedUserID(ctx, log)
	if err != nil {
		return
	}

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

	err = c.boardService.UpdateBoard(userID, boardID, request.Title, request.Description)
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

// DeleteBoard godoc
// @Summary      Delete a board
// @Description  Deletes a board by its UUID
// @Tags         boards
// @Produce      json
// @Param        id   path      string  true  "Board UUID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /boards/{id} [delete]
func (c *boardController) DeleteBoard(ctx *gin.Context) {
	const fn = "adapters.controller.DeleteBoard"
	log := slog.With(slog.String("fn", fn))

	userID, err := getAuthorizedUserID(ctx, log)
	if err != nil {
		return
	}

	boardIDParam := ctx.Param("id")
	boardID, err := uuid.Parse(boardIDParam)
	if err != nil {
		log.Error("invalid board ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	err = c.boardService.DeleteBoard(userID, boardID)
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

func getAuthorizedUserID(ctx *gin.Context, log *slog.Logger) (uuid.UUID, error) {
	userIDValue, exists := ctx.Get(UserIDKey)
	if !exists {
		log.Error("user ID not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return uuid.Nil, errors.New("user not authenticated")
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		log.Error("invalid user ID type in context")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return uuid.Nil, errors.New("invalid user ID")
	}

	return userID, nil
}

// Helper function to convert entity to response DTO
func toBoardResponse(board *entity.Board) boardResponse {
	return boardResponse{
		ID:          board.ID.String(),
		OwnerID:     board.OwnerID.String(),
		Title:       board.Title,
		Description: board.Description,
		CreatedAt:   board.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   board.UpdatedAt.Format(time.RFC3339),
	}
}
