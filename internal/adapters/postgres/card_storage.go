package storage

import (
	"context"
	"errors"
	"student-kanban/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type cardRepository struct {
	db *gorm.DB
}

func NewCardRepository(db *gorm.DB) *cardRepository {
	return &cardRepository{db: db}
}

func (r *cardRepository) Create(ctx context.Context, c *entity.Card) error {
	return r.db.WithContext(ctx).Create(c).Error
}

func (r *cardRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Card, error) {
	var c entity.Card
	err := r.db.WithContext(ctx).
		Where("id = ? AND is_archived = false", id).
		First(&c).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrCardNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *cardRepository) GetAllByListID(ctx context.Context, listID uuid.UUID) ([]*entity.Card, error) {
	var cards []*entity.Card
	err := r.db.WithContext(ctx).
		Where("list_id = ?", listID).
		Order("position ASC").
		Find(&cards).Error
	return cards, err
}

func (r *cardRepository) Update(ctx context.Context, c *entity.Card) error {
	return r.db.WithContext(ctx).Save(c).Error
}

func (r *cardRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&entity.Card{}, "id = ?", id).Error
}

func (r *cardRepository) GetMaxPosition(ctx context.Context, listID uuid.UUID) (float64, error) {
	var maxPos float64
	err := r.db.WithContext(ctx).
		Model(&entity.Card{}).
		Where("list_id = ?", listID).
		Select("COALESCE(MAX(position), 0)").
		Scan(&maxPos).Error
	return maxPos, err
}
