package services

import (
	"context"
	"student-kanban/internal/domain/entity"

	"github.com/google/uuid"
)

type CardRepository interface {
	Create(ctx context.Context, card *entity.Card) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Card, error)
	GetAllByListID(ctx context.Context, listID uuid.UUID) ([]entity.Card, error)
	Update(ctx context.Context, card *entity.Card) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetMaxPosition(ctx context.Context, listID uuid.UUID) (float64, error)
}

type listRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.List, error)
}

type cardService struct {
	cardRepo CardRepository
	listRepo listRepository
}

func NewCardService(cardRepo CardRepository, listRepo listRepository) *cardService {
	return &cardService{
		cardRepo: cardRepo,
		listRepo: listRepo,
	}
}
