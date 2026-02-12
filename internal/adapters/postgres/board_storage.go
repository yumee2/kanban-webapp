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

func NewboardStorage(db *gorm.DB) *boardStorage {
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
	result := s.db.Preload("Owner").Preload("Lists").Where("id = ?", id).First(&board)

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

// GetAllBoards retrieves all boards with pagination support
func (s *boardStorage) GetAllBoards(limit, offset int) ([]*entity.Board, error) {
	const fn = "adapters.repository.GetAllBoards"

	var boards []*entity.Board
	query := s.db.Preload("Owner").Preload("Lists")

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	result := query.Find(&boards)
	if result.Error != nil {
		return nil, fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	return boards, nil
}

// UpdateBoard updates a board's fields
func (s *boardStorage) UpdateBoard(board *entity.Board) error {
	const fn = "adapters.repository.UpdateBoard"

	result := s.db.Model(&entity.Board{}).Where("id = ?", board.ID).Updates(board)

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

// DeleteBoardsByOwnerID deletes all boards for a specific owner
func (s *boardStorage) DeleteBoardsByOwnerID(ownerID uuid.UUID) error {
	const fn = "adapters.repository.DeleteBoardsByOwnerID"

	result := s.db.Where("owner_id = ?", ownerID).Delete(&entity.Board{})

	if result.Error != nil {
		return fmt.Errorf("%s: %w", fn, result.Error)
	}

	return nil
}

// BoardExists checks if a board exists by ID
func (s *boardStorage) BoardExists(id uuid.UUID) (bool, error) {
	const fn = "adapters.repository.BoardExists"

	var count int64
	result := s.db.Model(&entity.Board{}).Where("id = ?", id).Count(&count)

	if result.Error != nil {
		return false, fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	return count > 0, nil
}

// GetBoardWithLists retrieves a board with its lists only (no owner preload)
func (s *boardStorage) GetBoardWithLists(id uuid.UUID) (*entity.Board, error) {
	const fn = "adapters.repository.GetBoardWithLists"

	var board entity.Board
	result := s.db.Preload("Lists").Where("id = ?", id).First(&board)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return &entity.Board{}, entity.ErrBoardNotFound
		}
		return &entity.Board{}, fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	return &board, nil
}

// CountBoardsByOwner counts the number of boards for a specific owner
func (s *boardStorage) CountBoardsByOwner(ownerID uuid.UUID) (int64, error) {
	const fn = "adapters.repository.CountBoardsByOwner"

	var count int64
	result := s.db.Model(&entity.Board{}).Where("owner_id = ?", ownerID).Count(&count)

	if result.Error != nil {
		return 0, fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	return count, nil
}

// CountAllBoards counts total number of boards
func (s *boardStorage) CountAllBoards() (int64, error) {
	const fn = "adapters.repository.CountAllBoards"

	var count int64
	result := s.db.Model(&entity.Board{}).Count(&count)

	if result.Error != nil {
		return 0, fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	return count, nil
}

// CheckBoardOwnership verifies if a user owns a specific board
func (s *boardStorage) CheckBoardOwnership(boardID, ownerID uuid.UUID) (bool, error) {
	const fn = "adapters.repository.CheckBoardOwnership"

	var count int64
	result := s.db.Model(&entity.Board{}).Where("id = ? AND owner_id = ?", boardID, ownerID).Count(&count)

	if result.Error != nil {
		return false, fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	return count > 0, nil
}
