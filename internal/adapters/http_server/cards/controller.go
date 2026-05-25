package cardshttp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"student-kanban/internal/adapters/http_server/middleware"
	"student-kanban/internal/domain/entity"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CardService interface {
	CreateCard(ctx context.Context, userID uuid.UUID, listID uuid.UUID, title string, description *string, dueDate *time.Time, priority entity.CardPriority, isFavorite bool) (*entity.Card, error)
	GetCardByID(ctx context.Context, id uuid.UUID) (*entity.Card, error)
	GetCardsByListID(ctx context.Context, listID uuid.UUID) ([]*entity.Card, error)
	UpdateCard(ctx context.Context, id uuid.UUID, title *string, description *string, dueDate *time.Time, priority *entity.CardPriority, position *float64, isFavorite *bool, isArchived *bool) (*entity.Card, error)
	DeleteCard(ctx context.Context, id uuid.UUID) error
	MoveCard(ctx context.Context, id uuid.UUID, position float64, listID *uuid.UUID) (*entity.Card, error)
}

type cardController struct {
	cardService CardService
}

func NewCardController(cardService CardService) *cardController {
	return &cardController{
		cardService: cardService,
	}
}

// CreateCard godoc
// @Summary      Create a new card
// @Description  Creates a new card inside a list
// @Tags         cards
// @Accept       json
// @Produce      json
// @Param        body  body      createCardRequest  true  "Card data"
// @Success      201   {object}  CardResponse
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /cards [post]
func (c *cardController) CreateCard(ctx *gin.Context) {
	const fn = "adapters.controller.CreateCard"
	log := slog.With(slog.String("fn", fn))

	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var request createCardRequest
	if err := ctx.BindJSON(&request); err != nil {
		log.Error("failed to parse json body", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card, err := c.cardService.CreateCard(ctx.Request.Context(), userID, request.ListID, request.Title, request.Description, request.DueDate, request.Priority, request.IsFavorite)
	if err != nil {
		if errors.Is(err, entity.ErrInvalidListID) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid list ID"})
			return
		}
		if errors.Is(err, entity.ErrListNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "List not found"})
			return
		}
		log.Error("failed to create card", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create card"})
		return
	}

	ctx.JSON(http.StatusCreated, ToCardResponse(card))
}

// GetCard godoc
// @Summary      Get a card by ID
// @Description  Returns a single card by its UUID
// @Tags         cards
// @Produce      json
// @Param        id   path      string  true  "Card UUID"
// @Success      200  {object}  CardResponse
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /cards/{id} [get]
func (c *cardController) GetCard(ctx *gin.Context) {
	const fn = "adapters.controller.GetCard"
	log := slog.With(slog.String("fn", fn))

	cardID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		log.Error("invalid card ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card ID"})
		return
	}

	card, err := c.cardService.GetCardByID(ctx.Request.Context(), cardID)
	if err != nil {
		if errors.Is(err, entity.ErrCardNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Card not found"})
			return
		}
		log.Error("failed to get card", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve card"})
		return
	}

	ctx.JSON(http.StatusOK, ToCardResponse(card))
}

// GetCardsByList godoc
// @Summary      Get all cards for a list
// @Description  Returns all cards belonging to a given list
// @Tags         cards
// @Produce      json
// @Param        list_id  query     string  true  "List UUID"
// @Success      200      {array}   CardResponse
// @Failure      400      {object}  map[string]string
// @Failure      500      {object}  map[string]string
// @Security     BearerAuth
// @Router       /cards [get]
func (c *cardController) GetCardsByList(ctx *gin.Context) {
	const fn = "adapters.controller.GetCardsByList"
	log := slog.With(slog.String("fn", fn))

	listID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		log.Error("invalid list ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid list ID"})
		return
	}

	cards, err := c.cardService.GetCardsByListID(ctx.Request.Context(), listID)
	if err != nil {
		log.Error("failed to get cards", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve cards"})
		return
	}

	response := make([]CardResponse, len(cards))
	for i, card := range cards {
		response[i] = ToCardResponse(card)
	}

	ctx.JSON(http.StatusOK, response)
}

// UpdateCard godoc
// @Summary      Update a card
// @Description  Updates a card's fields. At least one field must be provided
// @Tags         cards
// @Accept       json
// @Produce      json
// @Param        id    path      string            true  "Card UUID"
// @Param        body  body      updateCardRequest  true  "Fields to update"
// @Success      200   {object}  CardResponse
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /cards/{id} [put]
func (c *cardController) UpdateCard(ctx *gin.Context) {
	const fn = "adapters.controller.UpdateCard"
	log := slog.With(slog.String("fn", fn))

	cardID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		log.Error("invalid card ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card ID"})
		return
	}

	var request updateCardRequest
	if err := ctx.BindJSON(&request); err != nil {
		log.Error("failed to parse json body", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Title == nil && request.Description == nil && request.DueDate == nil && request.Priority == nil && request.Position == nil && request.IsFavorite == nil && request.IsArchived == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "At least one field must be provided"})
		return
	}

	card, err := c.cardService.UpdateCard(ctx.Request.Context(), cardID, request.Title, request.Description, request.DueDate, request.Priority, request.Position, request.IsFavorite, request.IsArchived)
	if err != nil {
		if errors.Is(err, entity.ErrCardNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Card not found"})
			return
		}
		if errors.Is(err, entity.ErrInvalidCardTitle) || errors.Is(err, entity.ErrInvalidCardPriority) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Error("failed to update card", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update card"})
		return
	}

	ctx.JSON(http.StatusOK, ToCardResponse(card))
}

// DeleteCard godoc
// @Summary      Delete a card
// @Description  Deletes a card by its UUID
// @Tags         cards
// @Produce      json
// @Param        id   path      string  true  "Card UUID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Security     BearerAuth
// @Router       /cards/{id} [delete]
func (c *cardController) DeleteCard(ctx *gin.Context) {
	const fn = "adapters.controller.DeleteCard"
	log := slog.With(slog.String("fn", fn))

	cardID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		log.Error("invalid card ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card ID"})
		return
	}

	err = c.cardService.DeleteCard(ctx.Request.Context(), cardID)
	if err != nil {
		if errors.Is(err, entity.ErrCardNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Card not found"})
			return
		}
		log.Error("failed to delete card", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete card"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Card deleted successfully"})
}

// MoveCard godoc
// @Summary      Move a card
// @Description  Updates a card's position, optionally moving it to a different list
// @Tags         cards
// @Accept       json
// @Produce      json
// @Param        id    path      string          true  "Card UUID"
// @Param        body  body      moveCardRequest  true  "Move data"
// @Success      200   {object}  CardResponse
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Security     BearerAuth
// @Router       /cards/{id}/move [patch]
func (c *cardController) MoveCard(ctx *gin.Context) {
	const fn = "adapters.controller.MoveCard"
	log := slog.With(slog.String("fn", fn))

	cardID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		log.Error("invalid card ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card ID"})
		return
	}

	var request moveCardRequest
	if err := ctx.BindJSON(&request); err != nil {
		log.Error("failed to parse json body", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card, err := c.cardService.MoveCard(ctx.Request.Context(), cardID, request.Position, request.ListID)
	if err != nil {
		if errors.Is(err, entity.ErrCardNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Card not found"})
			return
		}
		if errors.Is(err, entity.ErrListNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Target list not found"})
			return
		}
		log.Error("failed to move card", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move card"})
		return
	}

	ctx.JSON(http.StatusOK, ToCardResponse(card))
}

func ToCardResponse(card *entity.Card) CardResponse {
	return CardResponse{
		ID:          card.ID.String(),
		ListID:      card.ListID.String(),
		Title:       card.Title,
		Description: card.Description,
		DueDate:     card.DueDate,
		Priority:    card.Priority,
		Position:    card.Position,
		IsFavorite:  card.IsFavorite,
		IsArchived:  card.IsArchived,
	}
}
