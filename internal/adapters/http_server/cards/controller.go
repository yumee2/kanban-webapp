package cardshttp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"student-kanban/internal/domain/entity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CardService interface {
	CreateCard(ctx context.Context, listID uuid.UUID, title string, description *string, priority entity.CardPriority) (*entity.Card, error)
	GetCardByID(ctx context.Context, id uuid.UUID) (*entity.Card, error)
	GetCardsByListID(ctx context.Context, listID uuid.UUID) ([]*entity.Card, error)
	UpdateCard(ctx context.Context, id uuid.UUID, title *string, description *string, priority *entity.CardPriority, position *float64, isArchived *bool) (*entity.Card, error)
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

// CreateCard creates a new card inside a list
func (c *cardController) CreateCard(ctx *gin.Context) {
	const fn = "adapters.controller.CreateCard"
	log := slog.With(slog.String("fn", fn))

	var request createCardRequest
	if err := ctx.BindJSON(&request); err != nil {
		log.Error("failed to parse json body", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card, err := c.cardService.CreateCard(ctx.Request.Context(), request.ListID, request.Title, request.Description, request.Priority)
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

// GetCard retrieves a card by ID
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

// GetCardsByList retrieves all cards for a given list
func (c *cardController) GetCardsByList(ctx *gin.Context) {
	const fn = "adapters.controller.GetCardsByList"
	log := slog.With(slog.String("fn", fn))

	listID, err := uuid.Parse(ctx.Query("list_id"))
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

// UpdateCard updates a card's fields
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

	if request.Title == nil && request.Description == nil && request.Priority == nil && request.Position == nil && request.IsArchived == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "At least one field must be provided"})
		return
	}

	card, err := c.cardService.UpdateCard(ctx.Request.Context(), cardID, request.Title, request.Description, request.Priority, request.Position, request.IsArchived)
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

// DeleteCard deletes a card by ID
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

// MoveCard updates a card's position, optionally moving it to a different list
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
		Priority:    card.Priority,
		Position:    card.Position,
		IsArchived:  card.IsArchived,
	}
}
