package services

import (
	"context"
	"student-kanban/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type CreateCardDTO struct {
	ListID      uuid.UUID
	Title       string
	Description *string
	Priority    entity.CardPriority
}

type UpdateCardDTO struct {
	ID          uuid.UUID
	Title       *string
	Description *string
	Priority    *entity.CardPriority
	Position    *float64
	IsArchived  *bool
}

type CardDTO struct {
	ID          uuid.UUID
	ListID      uuid.UUID
	Title       string
	Description *string
	Priority    entity.CardPriority
	Position    float64
	IsArchived  bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

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

func (s *cardService) CreateCard(ctx context.Context, dto CreateCardDTO) (*CardDTO, error) {
	if dto.Title == "" {
		return nil, entity.ErrInvalidCardTitle
	}
	if dto.ListID == uuid.Nil {
		return nil, entity.ErrInvalidListID
	}

	// Set default priority if not provided
	if dto.Priority == "" {
		dto.Priority = entity.PriorityMedium
	}
	if !dto.Priority.IsValid() {
		return nil, entity.ErrInvalidCardPriority
	}

	// Verify list exists
	if _, err := s.listRepo.GetByID(ctx, dto.ListID); err != nil {
		return nil, err
	}

	// Auto-assign position
	maxPos, err := s.cardRepo.GetMaxPosition(ctx, dto.ListID)
	if err != nil {
		return nil, err
	}

	card := &entity.Card{
		ListID:      dto.ListID,
		Title:       dto.Title,
		Description: dto.Description,
		Priority:    dto.Priority,
		Position:    maxPos + 1,
	}

	if err := s.cardRepo.Create(ctx, card); err != nil {
		return nil, err
	}

	return toCardDTO(card), nil
}

func (s *cardService) GetCard(ctx context.Context, id uuid.UUID) (*CardDTO, error) {
	card, err := s.cardRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toCardDTO(card), nil
}

func (s *cardService) GetCardsByList(ctx context.Context, listID uuid.UUID) ([]CardDTO, error) {
	if listID == uuid.Nil {
		return nil, entity.ErrInvalidListID
	}

	cards, err := s.cardRepo.GetAllByListID(ctx, listID)
	if err != nil {
		return nil, err
	}

	dtos := make([]CardDTO, len(cards))
	for i, c := range cards {
		dtos[i] = *toCardDTO(&c)
	}
	return dtos, nil
}

func (s *cardService) UpdateCard(ctx context.Context, dto UpdateCardDTO) (*CardDTO, error) {
	if dto.ID == uuid.Nil {
		return nil, entity.ErrCardNotFound
	}

	card, err := s.cardRepo.GetByID(ctx, dto.ID)
	if err != nil {
		return nil, err
	}

	// Apply partial updates — only patch what was provided
	if dto.Title != nil {
		if *dto.Title == "" {
			return nil, entity.ErrInvalidCardTitle
		}
		card.Title = *dto.Title
	}
	if dto.Description != nil {
		card.Description = dto.Description
	}
	if dto.Priority != nil {
		if !dto.Priority.IsValid() {
			return nil, entity.ErrInvalidCardPriority
		}
		card.Priority = *dto.Priority
	}
	if dto.Position != nil {
		card.Position = *dto.Position
	}
	if dto.IsArchived != nil {
		card.IsArchived = *dto.IsArchived
	}

	if err := s.cardRepo.Update(ctx, card); err != nil {
		return nil, err
	}

	return toCardDTO(card), nil
}

func (s *cardService) DeleteCard(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return entity.ErrCardNotFound
	}
	// Verify it exists before deleting
	if _, err := s.cardRepo.GetByID(ctx, id); err != nil {
		return err
	}
	return s.cardRepo.Delete(ctx, id)
}

func toCardDTO(c *entity.Card) *CardDTO {
	return &CardDTO{
		ID:          c.ID,
		ListID:      c.ListID,
		Title:       c.Title,
		Description: c.Description,
		Priority:    c.Priority,
		Position:    c.Position,
		IsArchived:  c.IsArchived,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}
