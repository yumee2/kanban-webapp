package storage

import (
	"errors"
	"fmt"
	"student-kanban/internal/domain/entity"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type boardStorage struct {
	db *gorm.DB
}

func NewBoardStorage(db *gorm.DB) *boardStorage {
	return &boardStorage{db: db}
}

// CreateBoard creates a new board
func (s *boardStorage) CreateBoard(board *entity.Board) (string, error) {
	const fn = "adapters.repository.CreateBoard"

	result := s.db.Create(&board)
	if result.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) {
			// Handle potential unique constraint violations if you add any
			if pgErr.Code == "23505" {
				return "", fmt.Errorf("%s: duplicate board constraint violated", fn)
			}
			// Handle foreign key violations
			if pgErr.Code == "23503" {
				return "", entity.ErrOwnerNotFound
			}
		}
		return "", fmt.Errorf("%s: %w", fn, result.Error)
	}

	return board.ID.String(), nil
}

// GetBoardByID retrieves a board by its ID
func (s *boardStorage) GetBoardByID(id uuid.UUID) (*entity.Board, error) {
	const fn = "adapters.repository.GetBoardByID"

	var board entity.Board
	result := s.db.Preload("Owner").Preload("Tags").Preload("Lists").Where("id = ?", id).First(&board)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return &entity.Board{}, entity.ErrBoardNotFound
		}
		return &entity.Board{}, fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	return &board, nil
}

// GetBoardsByOwnerID retrieves all boards for a specific owner
func (s *boardStorage) GetBoardsByOwnerID(ownerID uuid.UUID) ([]*entity.Board, error) {
	const fn = "adapters.repository.GetBoardsByOwnerID"

	var boards []*entity.Board
	result := s.db.Preload("Owner").Preload("Lists").Where("owner_id = ?", ownerID).Find(&boards)

	if result.Error != nil {
		return nil, fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	return boards, nil
}

// UpdateBoardFields updates specific fields of a board using a map
func (s *boardStorage) UpdateBoardFields(id uuid.UUID, updates map[string]interface{}) error {
	const fn = "adapters.repository.UpdateBoardFields"

	result := s.db.Model(&entity.Board{}).Where("id = ?", id).Updates(updates)

	if result.Error != nil {
		var pgErr *pgconn.PgError
		if errors.As(result.Error, &pgErr) {
			// Handle foreign key violations
			if pgErr.Code == "23503" {
				return entity.ErrOwnerNotFound
			}
		}
		return fmt.Errorf("%s: %w", fn, result.Error)
	}

	if result.RowsAffected == 0 {
		return entity.ErrBoardNotFound
	}

	return nil
}

// DeleteBoard deletes a board by ID
func (s *boardStorage) DeleteBoard(id uuid.UUID) error {
	const fn = "adapters.repository.DeleteBoard"

	result := s.db.Delete(&entity.Board{}, "id = ?", id)

	if result.Error != nil {
		return fmt.Errorf("%s: %w", fn, result.Error)
	}

	if result.RowsAffected == 0 {
		return entity.ErrBoardNotFound
	}

	return nil
}
