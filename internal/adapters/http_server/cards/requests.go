package cardshttp

import (
	"student-kanban/internal/domain/entity"

	"github.com/google/uuid"
)

type createCardRequest struct {
	ListID      uuid.UUID           `json:"list_id" binding:"required"`
	Title       string              `json:"title" binding:"required"`
	Description *string             `json:"description"`
	Priority    entity.CardPriority `json:"priority"`
}

type updateCardRequest struct {
	Title       *string              `json:"title"`
	Description *string              `json:"description"`
	Priority    *entity.CardPriority `json:"priority"`
	Position    *float64             `json:"position"`
	IsArchived  *bool                `json:"is_archived"`
}

type moveCardRequest struct {
	Position float64    `json:"position" binding:"required"`
	ListID   *uuid.UUID `json:"list_id"`
}

type CardResponse struct {
	ID          string              `json:"id"`
	ListID      string              `json:"list_id"`
	Title       string              `json:"title"`
	Description *string             `json:"description"`
	Priority    entity.CardPriority `json:"priority"`
	Position    float64             `json:"position"`
	IsArchived  bool                `json:"is_archived"`
}
