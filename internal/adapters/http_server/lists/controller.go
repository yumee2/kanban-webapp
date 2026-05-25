package listshttp

import (
	"context"
	"net/http"
	"student-kanban/internal/domain/entity"
	"student-kanban/internal/domain/services"
	"time"

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
	DueDate     *time.Time          `json:"due_date"`
	Priority    entity.CardPriority `json:"priority"`
	Position    float64             `json:"position"`
	IsFavorite  bool                `json:"is_favorite"`
	IsArchived  bool                `json:"is_archived"`
	Tags        []tagResponse       `json:"tags"`
}

type tagResponse struct {
	ID      string `json:"id"`
	BoardID string `json:"board_id"`
	Name    string `json:"name"`
	Color   string `json:"color"`
}

type ListHandler struct {
	listService ListService
}

func NewListController(listService ListService) *ListHandler {
	return &ListHandler{
		listService: listService,
	}
}

// CreateList godoc
// @Summary      Create a new list
// @Description  Creates a new list inside a board
// @Tags         lists
// @Accept       json
// @Produce      json
// @Param        body  body      services.CreateListDTO  true  "List data"
// @Success      201   {object}  services.ListDTO
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /lists [post]
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

// GetList godoc
// @Summary      Get a list by ID
// @Description  Returns a single list with its cards
// @Tags         lists
// @Produce      json
// @Param        id   path      string  true  "List UUID"
// @Success      200  {object}  listResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /lists/{id} [get]
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

// GetListsByBoard godoc
// @Summary      Get all lists for a board
// @Description  Returns all lists belonging to a given board, each with their cards
// @Tags         lists
// @Produce      json
// @Param        id   path      string  true  "Board UUID"
// @Success      200  {array}   listResponse
// @Failure      400  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /boards/{id}/lists [get]
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

	listResponses := make([]listResponse, 0, len(lists))

	for _, list := range lists {
		listResponses = append(listResponses, toListResponse(list))
	}

	c.JSON(http.StatusOK, listResponses)
}

// UpdateList godoc
// @Summary      Update a list
// @Description  Updates a list's title and/or position
// @Tags         lists
// @Accept       json
// @Produce      json
// @Param        id    path      string                  true  "List UUID"
// @Param        body  body      services.UpdateListDTO  true  "Fields to update"
// @Success      200   {object}  listResponse
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /lists/{id} [put]
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

// DeleteList godoc
// @Summary      Delete a list
// @Description  Deletes a list and all its cards by UUID
// @Tags         lists
// @Produce      json
// @Param        id   path      string  true  "List UUID"
// @Success      204
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /lists/{id} [delete]
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

// UpdatePosition godoc
// @Summary      Update list position
// @Description  Updates the position of a list within its board
// @Tags         lists
// @Accept       json
// @Produce      json
// @Param        id    path      string  true  "List UUID"
// @Param        body  body      object  true  "Position data"  example({"position": 1.5})
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /lists/{id}/position [patch]
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
		tags := make([]tagResponse, len(card.Tags))
		for j, tag := range card.Tags {
			tags[j] = tagResponse{
				ID:      tag.ID.String(),
				BoardID: tag.BoardID.String(),
				Name:    tag.Name,
				Color:   tag.Color,
			}
		}

		cards[i] = cardResponse{
			ID:          card.ID.String(),
			ListID:      card.ListID.String(),
			Title:       card.Title,
			Description: card.Description,
			DueDate:     card.DueDate,
			Priority:    card.Priority,
			Position:    card.Position,
			IsFavorite:  card.IsFavorite,
			IsArchived:  card.IsArchived,
			Tags:        tags,
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
