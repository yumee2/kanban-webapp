package storage

import (
	"context"
	"errors"
	"student-kanban/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type tagRepository struct {
	db *gorm.DB
}

func NewTagRepository(db *gorm.DB) *tagRepository {
	return &tagRepository{db: db}
}

func (r *tagRepository) Create(ctx context.Context, tag *entity.Tag) error {
	return r.db.WithContext(ctx).Create(tag).Error
}

func (r *tagRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error) {
	var tag entity.Tag
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&tag).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrTagNotFound
		}
		return nil, err
	}
	return &tag, nil
}

func (r *tagRepository) GetAllByBoardID(ctx context.Context, boardID uuid.UUID) ([]*entity.Tag, error) {
	var tags []*entity.Tag
	err := r.db.WithContext(ctx).
		Where("board_id = ?", boardID).
		Find(&tags).Error
	return tags, err
}

func (r *tagRepository) Update(ctx context.Context, tag *entity.Tag) error {
	return r.db.WithContext(ctx).Save(tag).Error
}

func (r *tagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Delete(&entity.Tag{}, "id = ?", id).Error
}

func (r *tagRepository) AttachToCard(ctx context.Context, cardID uuid.UUID, tagID uuid.UUID) error {
	card := &entity.Card{}
	card.ID = cardID

	tag := &entity.Tag{}
	tag.ID = tagID

	err := r.db.WithContext(ctx).
		Model(card).
		Association("Tags").
		Append(tag)
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return entity.ErrTagAlreadyAttached
		}
		return err
	}
	return nil
}

func (r *tagRepository) DetachFromCard(ctx context.Context, cardID uuid.UUID, tagID uuid.UUID) error {
	card := &entity.Card{}
	card.ID = cardID

	tag := &entity.Tag{}
	tag.ID = tagID

	return r.db.WithContext(ctx).
		Model(card).
		Association("Tags").
		Delete(tag)
}

func (r *tagRepository) GetCardTags(ctx context.Context, cardID uuid.UUID) ([]*entity.Tag, error) {
	var card entity.Card
	err := r.db.WithContext(ctx).
		Preload("Tags").
		Where("id = ?", cardID).
		First(&card).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrCardNotFound
		}
		return nil, err
	}
	return func() []*entity.Tag {
		tags := make([]*entity.Tag, len(card.Tags))
		for i := range card.Tags {
			tags[i] = &card.Tags[i]
		}
		return tags
	}(), nil
}
