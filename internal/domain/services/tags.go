package services

import (
	"context"
	"student-kanban/internal/domain/entity"

	"github.com/google/uuid"
)

type TagRepository interface {
	Create(ctx context.Context, tag *entity.Tag) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error)
	GetAllByBoardID(ctx context.Context, boardID uuid.UUID) ([]*entity.Tag, error)
	Update(ctx context.Context, tag *entity.Tag) error
	Delete(ctx context.Context, id uuid.UUID) error
	AttachToCard(ctx context.Context, cardID uuid.UUID, tagID uuid.UUID) error
	DetachFromCard(ctx context.Context, cardID uuid.UUID, tagID uuid.UUID) error
	GetCardTags(ctx context.Context, cardID uuid.UUID) ([]*entity.Tag, error)
}

type boardRepository interface {
	GetBoardByID(id uuid.UUID) (*entity.Board, error)
}

type cardRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Card, error)
}

type tagService struct {
	tagRepo   TagRepository
	boardRepo boardRepository
	cardRepo  cardRepository
}

func NewTagService(tagRepo TagRepository, boardRepo boardRepository, cardRepo cardRepository) *tagService {
	return &tagService{
		tagRepo:   tagRepo,
		boardRepo: boardRepo,
		cardRepo:  cardRepo,
	}
}

func (s *tagService) CreateTag(ctx context.Context, boardID uuid.UUID, name, color string) (*entity.Tag, error) {
	if name == "" {
		return nil, entity.ErrInvalidTagName
	}
	if color == "" {
		return nil, entity.ErrInvalidTagColor
	}

	if _, err := s.boardRepo.GetBoardByID(boardID); err != nil {
		return nil, err
	}

	tag := &entity.Tag{
		BoardID: boardID,
		Name:    name,
		Color:   color,
	}

	if err := s.tagRepo.Create(ctx, tag); err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *tagService) GetTagByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error) {
	return s.tagRepo.GetByID(ctx, id)
}

func (s *tagService) GetTagsByBoard(ctx context.Context, boardID uuid.UUID) ([]*entity.Tag, error) {
	return s.tagRepo.GetAllByBoardID(ctx, boardID)
}

func (s *tagService) UpdateTag(ctx context.Context, id uuid.UUID, name, color *string) (*entity.Tag, error) {
	tag, err := s.tagRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		if *name == "" {
			return nil, entity.ErrInvalidTagName
		}
		tag.Name = *name
	}
	if color != nil {
		if *color == "" {
			return nil, entity.ErrInvalidTagColor
		}
		tag.Color = *color
	}

	if err := s.tagRepo.Update(ctx, tag); err != nil {
		return nil, err
	}

	return tag, nil
}

func (s *tagService) DeleteTag(ctx context.Context, id uuid.UUID) error {
	if _, err := s.tagRepo.GetByID(ctx, id); err != nil {
		return err
	}
	return s.tagRepo.Delete(ctx, id)
}

func (s *tagService) AttachTag(ctx context.Context, cardID, tagID uuid.UUID) error {
	if _, err := s.cardRepo.GetByID(ctx, cardID); err != nil {
		return err
	}
	if _, err := s.tagRepo.GetByID(ctx, tagID); err != nil {
		return err
	}
	return s.tagRepo.AttachToCard(ctx, cardID, tagID)
}

func (s *tagService) DetachTag(ctx context.Context, cardID, tagID uuid.UUID) error {
	if _, err := s.cardRepo.GetByID(ctx, cardID); err != nil {
		return err
	}
	if _, err := s.tagRepo.GetByID(ctx, tagID); err != nil {
		return err
	}
	return s.tagRepo.DetachFromCard(ctx, cardID, tagID)
}

func (s *tagService) GetCardTags(ctx context.Context, cardID uuid.UUID) ([]*entity.Tag, error) {
	if _, err := s.cardRepo.GetByID(ctx, cardID); err != nil {
		return nil, err
	}
	return s.tagRepo.GetCardTags(ctx, cardID)
}
