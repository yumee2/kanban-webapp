package boardhttp

type createBoardRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type updateBoardRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
}

type boardResponse struct {
	ID          string  `json:"id"`
	OwnerID     string  `json:"owner_id"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type createBoardResponse struct {
	ID string `json:"id"`
}
