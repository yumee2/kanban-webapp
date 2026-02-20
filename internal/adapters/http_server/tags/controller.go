package tagshttp

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"student-kanban/internal/domain/entity"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TagService interface {
	CreateTag(ctx context.Context, boardID uuid.UUID, name, color string) (*entity.Tag, error)
	GetTagByID(ctx context.Context, id uuid.UUID) (*entity.Tag, error)
	GetTagsByBoard(ctx context.Context, boardID uuid.UUID) ([]*entity.Tag, error)
	UpdateTag(ctx context.Context, id uuid.UUID, name, color *string) (*entity.Tag, error)
	DeleteTag(ctx context.Context, id uuid.UUID) error
	AttachTag(ctx context.Context, cardID, tagID uuid.UUID) error
	DetachTag(ctx context.Context, cardID, tagID uuid.UUID) error
	GetCardTags(ctx context.Context, cardID uuid.UUID) ([]*entity.Tag, error)
}

type tagController struct {
	tagService TagService
}

func NewTagController(tagService TagService) *tagController {
	return &tagController{tagService: tagService}
}

func (c *tagController) CreateTag(ctx *gin.Context) {
	const fn = "adapters.controller.CreateTag"
	log := slog.With(slog.String("fn", fn))

	var request createTagRequest
	if err := ctx.BindJSON(&request); err != nil {
		log.Error("failed to parse json body", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tag, err := c.tagService.CreateTag(ctx.Request.Context(), request.BoardID, request.Name, request.Color)
	if err != nil {
		if errors.Is(err, entity.ErrBoardNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Board not found"})
			return
		}
		if errors.Is(err, entity.ErrInvalidTagName) || errors.Is(err, entity.ErrInvalidTagColor) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Error("failed to create tag", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tag"})
		return
	}

	ctx.JSON(http.StatusCreated, toTagResponse(tag))
}

func (c *tagController) GetTag(ctx *gin.Context) {
	const fn = "adapters.controller.GetTag"
	log := slog.With(slog.String("fn", fn))

	tagID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		log.Error("invalid tag ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
		return
	}

	tag, err := c.tagService.GetTagByID(ctx.Request.Context(), tagID)
	if err != nil {
		if errors.Is(err, entity.ErrTagNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
			return
		}
		log.Error("failed to get tag", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tag"})
		return
	}

	ctx.JSON(http.StatusOK, toTagResponse(tag))
}

func (c *tagController) GetTagsByBoard(ctx *gin.Context) {
	const fn = "adapters.controller.GetTagsByBoard"
	log := slog.With(slog.String("fn", fn))

	boardID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		log.Error("invalid board ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid board ID"})
		return
	}

	tags, err := c.tagService.GetTagsByBoard(ctx.Request.Context(), boardID)
	if err != nil {
		log.Error("failed to get tags", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tags"})
		return
	}

	response := make([]tagResponse, len(tags))
	for i, tag := range tags {
		response[i] = toTagResponse(tag)
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *tagController) UpdateTag(ctx *gin.Context) {
	const fn = "adapters.controller.UpdateTag"
	log := slog.With(slog.String("fn", fn))

	tagID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		log.Error("invalid tag ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
		return
	}

	var request updateTagRequest
	if err := ctx.BindJSON(&request); err != nil {
		log.Error("failed to parse json body", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if request.Name == nil && request.Color == nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "At least one field must be provided"})
		return
	}

	tag, err := c.tagService.UpdateTag(ctx.Request.Context(), tagID, request.Name, request.Color)
	if err != nil {
		if errors.Is(err, entity.ErrTagNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
			return
		}
		if errors.Is(err, entity.ErrInvalidTagName) || errors.Is(err, entity.ErrInvalidTagColor) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		log.Error("failed to update tag", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tag"})
		return
	}

	ctx.JSON(http.StatusOK, toTagResponse(tag))
}

func (c *tagController) DeleteTag(ctx *gin.Context) {
	const fn = "adapters.controller.DeleteTag"
	log := slog.With(slog.String("fn", fn))

	tagID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		log.Error("invalid tag ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
		return
	}

	if err := c.tagService.DeleteTag(ctx.Request.Context(), tagID); err != nil {
		if errors.Is(err, entity.ErrTagNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
			return
		}
		log.Error("failed to delete tag", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete tag"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Tag deleted successfully"})
}

func (c *tagController) AttachTag(ctx *gin.Context) {
	const fn = "adapters.controller.AttachTag"
	log := slog.With(slog.String("fn", fn))

	cardID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		log.Error("invalid card ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card ID"})
		return
	}

	tagID, err := uuid.Parse(ctx.Param("tag_id"))
	if err != nil {
		log.Error("invalid tag ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
		return
	}

	if err := c.tagService.AttachTag(ctx.Request.Context(), cardID, tagID); err != nil {
		if errors.Is(err, entity.ErrCardNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Card not found"})
			return
		}
		if errors.Is(err, entity.ErrTagNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
			return
		}
		if errors.Is(err, entity.ErrTagAlreadyAttached) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Tag already attached to card"})
			return
		}
		log.Error("failed to attach tag", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to attach tag"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Tag attached successfully"})
}

func (c *tagController) DetachTag(ctx *gin.Context) {
	const fn = "adapters.controller.DetachTag"
	log := slog.With(slog.String("fn", fn))

	cardID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		log.Error("invalid card ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card ID"})
		return
	}

	tagID, err := uuid.Parse(ctx.Param("tag_id"))
	if err != nil {
		log.Error("invalid tag ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tag ID"})
		return
	}

	if err := c.tagService.DetachTag(ctx.Request.Context(), cardID, tagID); err != nil {
		if errors.Is(err, entity.ErrCardNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Card not found"})
			return
		}
		if errors.Is(err, entity.ErrTagNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
			return
		}
		log.Error("failed to detach tag", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to detach tag"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Tag detached successfully"})
}

func (c *tagController) GetCardTags(ctx *gin.Context) {
	const fn = "adapters.controller.GetCardTags"
	log := slog.With(slog.String("fn", fn))

	cardID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		log.Error("invalid card ID", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid card ID"})
		return
	}

	tags, err := c.tagService.GetCardTags(ctx.Request.Context(), cardID)
	if err != nil {
		if errors.Is(err, entity.ErrCardNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Card not found"})
			return
		}
		log.Error("failed to get card tags", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tags"})
		return
	}

	response := make([]tagResponse, len(tags))
	for i, tag := range tags {
		response[i] = toTagResponse(tag)
	}

	ctx.JSON(http.StatusOK, response)
}

func toTagResponse(tag *entity.Tag) tagResponse {
	return tagResponse{
		ID:      tag.ID.String(),
		BoardID: tag.BoardID.String(),
		Name:    tag.Name,
		Color:   tag.Color,
	}
}
