package services

import (
	"context"
	"student-kanban/internal/domain/entity"
	"time"

	"github.com/google/uuid"
)

type ListRepository interface {
	Create(ctx context.Context, list *entity.List) error
	FindByID(ctx context.Context, id uuid.UUID) (*entity.List, error)
	FindByBoardID(ctx context.Context, boardID uuid.UUID) ([]entity.List, error)
	Update(ctx context.Context, list *entity.List) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdatePosition(ctx context.Context, id uuid.UUID, position float64) error
}

type CreateListDTO struct {
	BoardID  uuid.UUID `json:"board_id" binding:"required"`
	Title    string    `json:"title" binding:"required"`
	Position float64   `json:"position"`
}

type UpdateListDTO struct {
	Title    *string  `json:"title,omitempty"`
	Position *float64 `json:"position,omitempty"`
}

type ListDTO struct {
	ID        uuid.UUID `json:"id"`
	BoardID   uuid.UUID `json:"board_id"`
	Title     string    `json:"title"`
	Position  float64   `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type listService struct {
	listRepo  ListRepository
	boardRepo BoardRepository // To verify board exists
}

func NewListService(listRepo ListRepository, boardRepo BoardRepository) *listService {
	return &listService{
		listRepo:  listRepo,
		boardRepo: boardRepo,
	}
}

func (s *listService) CreateList(ctx context.Context, dto CreateListDTO) (*ListDTO, error) {
	// Validate input
	if dto.Title == "" {
		return nil, entity.ErrInvalidListTitle
	}

	if dto.BoardID == uuid.Nil {
		return nil, entity.ErrInvalidBoardID
	}

	// Verify board exists
	_, err := s.boardRepo.GetBoardByID(dto.BoardID)
	if err != nil {
		return nil, err
	}

	// Create list
	list := &entity.List{
		BoardID:  dto.BoardID,
		Title:    dto.Title,
		Position: dto.Position,
	}

	if err := s.listRepo.Create(ctx, list); err != nil {
		return nil, err
	}

	return toListDTO(list), nil
}

func (s *listService) GetListByID(ctx context.Context, id uuid.UUID) (*ListDTO, error) {
	list, err := s.listRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toListDTO(list), nil
}

func (s *listService) GetListsByBoardID(ctx context.Context, boardID uuid.UUID) ([]*ListDTO, error) {
	lists, err := s.listRepo.FindByBoardID(ctx, boardID)
	if err != nil {
		return nil, err
	}

	dtos := make([]*ListDTO, len(lists))
	for i, list := range lists {
		dtos[i] = toListDTO(&list)
	}

	return dtos, nil
}

func (s *listService) UpdateList(ctx context.Context, id uuid.UUID, dto UpdateListDTO) (*ListDTO, error) {
	// Get existing list
	list, err := s.listRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if dto.Title != nil {
		if *dto.Title == "" {
			return nil, entity.ErrInvalidListTitle
		}
		list.Title = *dto.Title
	}

	if dto.Position != nil {
		list.Position = *dto.Position
	}

	// Save updates
	if err := s.listRepo.Update(ctx, list); err != nil {
		return nil, err
	}

	// Fetch updated list
	updatedList, err := s.listRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return toListDTO(updatedList), nil
}

func (s *listService) DeleteList(ctx context.Context, id uuid.UUID) error {
	return s.listRepo.Delete(ctx, id)
}

func (s *listService) UpdateListPosition(ctx context.Context, id uuid.UUID, position float64) error {
	return s.listRepo.UpdatePosition(ctx, id, position)
}

// Helper function to convert List model to DTO
func toListDTO(list *entity.List) *ListDTO {
	return &ListDTO{
		ID:        list.ID,
		BoardID:   list.BoardID,
		Title:     list.Title,
		Position:  list.Position,
		CreatedAt: list.CreatedAt,
		UpdatedAt: list.UpdatedAt,
	}
}
