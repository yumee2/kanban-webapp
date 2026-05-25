package storage

import (
	"context"
	"errors"
	"student-kanban/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type listRepository struct {
	db *gorm.DB
}

func NewListRepository(db *gorm.DB) *listRepository {
	return &listRepository{db: db}
}

// Create creates a new list in the database
func (r *listRepository) Create(ctx context.Context, list *entity.List) error {
	if err := r.db.WithContext(ctx).Create(list).Error; err != nil {
		return err
	}
	return nil
}

// FindByID retrieves a list by its ID
func (r *listRepository) FindByID(ctx context.Context, id uuid.UUID) (*entity.List, error) {
	var list entity.List
	if err := r.db.WithContext(ctx).
		Where("id = ?", id).
		Preload("Cards").
		First(&list).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrListNotFound
		}
		return nil, err
	}
	return &list, nil
}

// FindByBoardID retrieves all lists belonging to a specific board
// Lists are ordered by position in ascending order
func (r *listRepository) FindByBoardID(ctx context.Context, boardID uuid.UUID) ([]entity.List, error) {
	var lists []entity.List
	if err := r.db.WithContext(ctx).
		Where("board_id = ?", boardID).
		Order("position ASC").
		Preload("Cards", func(db *gorm.DB) *gorm.DB {
			return db.Order("position ASC")
		}).
		Preload("Cards.Tags").
		Find(&lists).Error; err != nil {
		return nil, err
	}
	return lists, nil
}

// Update updates an existing list
func (r *listRepository) Update(ctx context.Context, list *entity.List) error {
	result := r.db.WithContext(ctx).
		Model(&entity.List{}).
		Where("id = ?", list.ID).
		Updates(map[string]interface{}{
			"title":      list.Title,
			"position":   list.Position,
			"updated_at": gorm.Expr("NOW()"),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return entity.ErrListNotFound
	}

	return nil
}

// Delete deletes a list by its ID
func (r *listRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var list entity.List
		if err := tx.Select("id").Where("id = ?", id).First(&list).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return entity.ErrListNotFound
			}
			return err
		}

		var cardIDs []uuid.UUID
		if err := tx.Model(&entity.Card{}).Where("list_id = ?", id).Pluck("id", &cardIDs).Error; err != nil {
			return err
		}

		if len(cardIDs) > 0 {
			if err := tx.Table("card_tags").Where("card_id IN ?", cardIDs).Delete(nil).Error; err != nil {
				return err
			}

			if err := tx.Where("id IN ?", cardIDs).Delete(&entity.Card{}).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("id = ?", id).Delete(&entity.List{}).Error; err != nil {
			return err
		}

		return nil
	})
}

// UpdatePosition updates only the position of a list
func (r *listRepository) UpdatePosition(ctx context.Context, id uuid.UUID, position float64) error {
	result := r.db.WithContext(ctx).
		Model(&entity.List{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"position":   position,
			"updated_at": gorm.Expr("NOW()"),
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return entity.ErrListNotFound
	}

	return nil
}

func (r *listRepository) GetMaxPosition(ctx context.Context, boardID uuid.UUID) (float64, error) {
	var maxPos float64
	err := r.db.WithContext(ctx).
		Model(&entity.List{}).
		Where("board_id = ?", boardID).
		Select("COALESCE(MAX(position), 0)").
		Scan(&maxPos).Error
	return maxPos, err
}
