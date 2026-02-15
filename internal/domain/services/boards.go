package services

import (
	"fmt"
	"student-kanban/internal/domain/entity"

	"github.com/google/uuid"
)

type BoardRepository interface {
	CreateBoard(board *entity.Board) (string, error)
	GetBoardByID(id uuid.UUID) (*entity.Board, error)
	GetBoardsByOwnerID(ownerID uuid.UUID) ([]*entity.Board, error)
	UpdateBoardFields(id uuid.UUID, updates map[string]interface{}) error
	DeleteBoard(id uuid.UUID) error
}

type boardService struct {
	repo BoardRepository
}

func NewBoardService(repo BoardRepository) *boardService {
	return &boardService{
		repo: repo,
	}
}

// CreateBoard creates a new board with the given parameters
func (s *boardService) CreateBoard(ownerID uuid.UUID, title, description string) (string, error) {
	const fn = "service.boardService.CreateBoard"

	// Validate input
	if title == "" {
		return "", fmt.Errorf("%s: title cannot be empty", fn)
	}

	if ownerID == uuid.Nil {
		return "", fmt.Errorf("%s: owner ID cannot be nil", fn)
	}

	// Create board entity
	board := &entity.Board{
		OwnerID:     ownerID,
		Title:       title,
		Description: &description,
	}

	boardID, err := s.repo.CreateBoard(board)
	if err != nil {
		return "", fmt.Errorf("%s: %w", fn, err)
	}

	return boardID, nil
}

// GetBoardByID retrieves a board by its ID
func (s *boardService) GetBoardByID(id uuid.UUID) (*entity.Board, error) {
	const fn = "service.boardService.GetBoardByID"

	if id == uuid.Nil {
		return nil, fmt.Errorf("%s: board ID cannot be nil", fn)
	}

	board, err := s.repo.GetBoardByID(id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return board, nil
}

// GetBoardsByOwnerID retrieves all boards for a specific owner
func (s *boardService) GetBoardsByOwnerID(ownerID uuid.UUID) ([]*entity.Board, error) {
	const fn = "service.boardService.GetBoardsByOwnerID"

	if ownerID == uuid.Nil {
		return nil, fmt.Errorf("%s: owner ID cannot be nil", fn)
	}

	boards, err := s.repo.GetBoardsByOwnerID(ownerID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return boards, nil
}

// UpdateBoard updates one or more fields of a board
func (s *boardService) UpdateBoard(id uuid.UUID, title *string, description *string) error {
	const fn = "service.boardService.UpdateBoard"

	if id == uuid.Nil {
		return fmt.Errorf("%s: board ID cannot be nil", fn)
	}

	updates := make(map[string]interface{})

	if title != nil {
		if *title == "" {
			return fmt.Errorf("%s: title cannot be empty", fn)
		}
		updates["title"] = *title
	}

	if description != nil {
		updates["description"] = *description
	}

	// At least one field must be updated
	if len(updates) == 0 {
		return fmt.Errorf("%s: no fields to update", fn)
	}

	err := s.repo.UpdateBoardFields(id, updates)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

// DeleteBoard deletes a board by ID
func (s *boardService) DeleteBoard(id uuid.UUID) error {
	const fn = "service.boardService.DeleteBoard"

	if id == uuid.Nil {
		return fmt.Errorf("%s: board ID cannot be nil", fn)
	}

	err := s.repo.DeleteBoard(id)
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}
