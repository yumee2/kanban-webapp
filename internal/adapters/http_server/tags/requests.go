package tagshttp

import "github.com/google/uuid"

type createTagRequest struct {
	BoardID uuid.UUID `json:"board_id" binding:"required"`
	Name    string    `json:"name" binding:"required"`
	Color   string    `json:"color" binding:"required"`
}

type updateTagRequest struct {
	Name  *string `json:"name"`
	Color *string `json:"color"`
}

type tagResponse struct {
	ID      string `json:"id"`
	BoardID string `json:"board_id"`
	Name    string `json:"name"`
	Color   string `json:"color"`
}
