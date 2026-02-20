package listshttp

import (
	"context"
	"net/http"
	"student-kanban/internal/domain/entity"
	"student-kanban/internal/domain/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ListService interface {
	CreateList(ctx context.Context, dto services.CreateListDTO) (*services.ListDTO, error)
	GetListByID(ctx context.Context, id uuid.UUID) (*services.ListDTO, error)
	GetListsByBoardID(ctx context.Context, boardID uuid.UUID) ([]*services.ListDTO, error)
	UpdateList(ctx context.Context, id uuid.UUID, dto services.UpdateListDTO) (*services.ListDTO, error)
	DeleteList(ctx context.Context, id uuid.UUID) error
	UpdateListPosition(ctx context.Context, id uuid.UUID, position float64) error
}

type cardResponse struct {
	ID          string              `json:"id"`
	ListID      string              `json:"list_id"`
	Title       string              `json:"title"`
	Description *string             `json:"description"`
	Priority    entity.CardPriority `json:"priority"`
	Position    float64             `json:"position"`
	IsArchived  bool                `json:"is_archived"`
}

type ListHandler struct {
	listService ListService
}

func NewListController(listService ListService) *ListHandler {
	return &ListHandler{
		listService: listService,
	}
}

func (h *ListHandler) CreateList(c *gin.Context) {
	var dto services.CreateListDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	list, err := h.listService.CreateList(c.Request.Context(), dto)
	if err != nil {
		switch err {
		case entity.ErrInvalidListTitle, entity.ErrInvalidBoardID:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case entity.ErrBoardNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "board not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create list"})
		}
		return
	}

	c.JSON(http.StatusCreated, list)
}

func (h *ListHandler) GetList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid list ID"})
		return
	}

	list, err := h.listService.GetListByID(c.Request.Context(), id)
	if err != nil {
		if err == entity.ErrListNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "list not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve list"})
		return
	}

	c.JSON(http.StatusOK, list)
}

func (h *ListHandler) GetListsByBoard(c *gin.Context) {
	idStr := c.Param("id")
	boardID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid board ID"})
		return
	}

	lists, err := h.listService.GetListsByBoardID(c.Request.Context(), boardID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve lists"})
		return
	}

	listResponses := make([]listResponse, len(lists))

	for _, list := range lists {
		listResponses = append(listResponses, toListResponse(list))
	}

	c.JSON(http.StatusOK, listResponses)
}

func (h *ListHandler) UpdateList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid list ID"})
		return
	}

	var dto services.UpdateListDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	list, err := h.listService.UpdateList(c.Request.Context(), id, dto)
	if err != nil {
		switch err {
		case entity.ErrListNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "list not found"})
		case entity.ErrInvalidListTitle:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update list"})
		}
		return
	}

	c.JSON(http.StatusOK, toListResponse(list))
}

func (h *ListHandler) DeleteList(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid list ID"})
		return
	}

	if err := h.listService.DeleteList(c.Request.Context(), id); err != nil {
		if err == entity.ErrListNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "list not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete list"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *ListHandler) UpdatePosition(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid list ID"})
		return
	}

	var body struct {
		Position float64 `json:"position" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.listService.UpdateListPosition(c.Request.Context(), id, body.Position); err != nil {
		if err == entity.ErrListNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "list not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update position"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "position updated successfully"})
}

type listResponse struct {
	ID       string         `json:"id"`
	BoardID  string         `json:"board_id"`
	Title    string         `json:"title"`
	Position float64        `json:"position"`
	Cards    []cardResponse `json:"cards"`
}

func toListResponse(list *services.ListDTO) listResponse {
	cards := make([]cardResponse, len(list.Cards))
	for i, card := range list.Cards {
		cards[i] = cardResponse{
			ID:          card.ID.String(),
			ListID:      card.ListID.String(),
			Title:       card.Title,
			Description: card.Description,
			Priority:    card.Priority,
			Position:    card.Position,
			IsArchived:  card.IsArchived,
		}
	}

	return listResponse{
		ID:       list.ID.String(),
		BoardID:  list.BoardID.String(),
		Title:    list.Title,
		Position: list.Position,
		Cards:    cards,
	}
}
