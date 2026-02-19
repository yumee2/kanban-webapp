package services

import (
	"context"
	"student-kanban/internal/domain/entity"

	"github.com/google/uuid"
)

type CardRepository interface {
	Create(ctx context.Context, card *entity.Card) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Card, error)
	GetAllByListID(ctx context.Context, listID uuid.UUID) ([]*entity.Card, error)
	Update(ctx context.Context, card *entity.Card) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetMaxPosition(ctx context.Context, listID uuid.UUID) (float64, error)
}

type listRepository interface {
	FindByID(ctx context.Context, id uuid.UUID) (*entity.List, error)
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

func (s *cardService) CreateCard(ctx context.Context, listID uuid.UUID, title string, description *string, priority entity.CardPriority) (*entity.Card, error) {
	if title == "" {
		return nil, entity.ErrInvalidCardTitle
	}
	if listID == uuid.Nil {
		return nil, entity.ErrInvalidListID
	}

	if priority == "" {
		priority = entity.PriorityMedium
	}
	if !priority.IsValid() {
		return nil, entity.ErrInvalidCardPriority
	}

	if _, err := s.listRepo.GetByID(ctx, listID); err != nil {
		return nil, err
	}

	maxPos, err := s.cardRepo.GetMaxPosition(ctx, listID)
	if err != nil {
		return nil, err
	}

	card := &entity.Card{
		ListID:      listID,
		Title:       title,
		Description: description,
		Priority:    priority,
		Position:    maxPos + 1,
	}

	if err := s.cardRepo.Create(ctx, card); err != nil {
		return nil, err
	}

	return card, nil
}

func (s *cardService) GetCardByID(ctx context.Context, id uuid.UUID) (*entity.Card, error) {
	return s.cardRepo.GetByID(ctx, id)
}

func (s *cardService) GetCardsByListID(ctx context.Context, listID uuid.UUID) ([]*entity.Card, error) {
	if listID == uuid.Nil {
		return nil, entity.ErrInvalidListID
	}
	return s.cardRepo.GetAllByListID(ctx, listID)
}

func (s *cardService) UpdateCard(ctx context.Context, id uuid.UUID, title *string, description *string, priority *entity.CardPriority, position *float64, isArchived *bool) (*entity.Card, error) {
	card, err := s.cardRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if title != nil {
		if *title == "" {
			return nil, entity.ErrInvalidCardTitle
		}
		card.Title = *title
	}
	if description != nil {
		card.Description = description
	}
	if priority != nil {
		if !priority.IsValid() {
			return nil, entity.ErrInvalidCardPriority
		}
		card.Priority = *priority
	}
	if position != nil {
		card.Position = *position
	}
	if isArchived != nil {
		card.IsArchived = *isArchived
	}

	if err := s.cardRepo.Update(ctx, card); err != nil {
		return nil, err
	}

	return card, nil
}

func (s *cardService) MoveCard(ctx context.Context, id uuid.UUID, position float64, listID *uuid.UUID) (*entity.Card, error) {
	card, err := s.cardRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// If moving to a different list, verify the target list exists
	if listID != nil && *listID != card.ListID {
		if _, err := s.listRepo.GetByID(ctx, *listID); err != nil {
			return nil, err
		}
		card.ListID = *listID
	}

	card.Position = position

	if err := s.cardRepo.Update(ctx, card); err != nil {
		return nil, err
	}

	return card, nil
}

func (s *cardService) DeleteCard(ctx context.Context, id uuid.UUID) error {
	if _, err := s.cardRepo.GetByID(ctx, id); err != nil {
		return err
	}
	return s.cardRepo.Delete(ctx, id)
}
