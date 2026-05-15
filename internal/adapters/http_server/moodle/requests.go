package moodlehttp

type connectRequest struct {
	BaseURL  string `json:"base_url" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Service  string `json:"service"`
}

type importBoardRequest struct {
	CourseID int64 `json:"course_id" binding:"required"`
}

type importBoardResponse struct {
	BoardID string `json:"board_id"`
}
