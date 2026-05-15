package moodlehttp

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"student-kanban/internal/adapters/http_server/middleware"
	"student-kanban/internal/adapters/moodle"
	"student-kanban/internal/domain/entity"
	"student-kanban/internal/domain/services"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MoodleService interface {
	Connect(userID uuid.UUID, input services.ConnectMoodleInput) (*services.MoodleConnectionInfo, error)
	GetConnection(userID uuid.UUID) (*services.MoodleConnectionInfo, error)
	GetCourses(userID uuid.UUID) ([]moodle.Course, error)
	ImportCourseAsBoard(ctx context.Context, userID uuid.UUID, courseID int64) (string, error)
}

type Controller struct {
	service MoodleService
}

func NewController(service MoodleService) *Controller {
	return &Controller{service: service}
}

// Connect godoc
// @Summary      Connect a Moodle account
// @Description  Authenticates against Moodle with user credentials, exchanges them for a token, and stores the connection for the authenticated user
// @Tags         moodle
// @Accept       json
// @Produce      json
// @Param        body  body      connectRequest                true  "Moodle credentials"
// @Success      200   {object}  services.MoodleConnectionInfo
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Failure      502   {object}  map[string]string
// @Security     BearerAuth
// @Router       /integrations/moodle/connect [post]
func (c *Controller) Connect(ctx *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var request connectRequest
	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	connection, err := c.service.Connect(userID, services.ConnectMoodleInput{
		BaseURL:  request.BaseURL,
		Username: request.Username,
		Password: request.Password,
		Service:  request.Service,
	})
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrMoodleAuthFailed):
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Moodle authentication failed"})
		case errors.Is(err, entity.ErrMoodleTokenKeyMissing):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Server Moodle encryption is not configured"})
		default:
			if strings.Contains(err.Error(), "base URL") || strings.Contains(err.Error(), "required") {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, connection)
}

// GetConnection godoc
// @Summary      Get saved Moodle connection
// @Description  Returns the authenticated user's saved Moodle connection details
// @Tags         moodle
// @Produce      json
// @Success      200  {object}  services.MoodleConnectionInfo
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /integrations/moodle/connection [get]
func (c *Controller) GetConnection(ctx *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	connection, err := c.service.GetConnection(userID)
	if err != nil {
		if errors.Is(err, entity.ErrMoodleConnectionNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Moodle account is not connected"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load Moodle connection"})
		return
	}

	ctx.JSON(http.StatusOK, connection)
}

// ListCourses godoc
// @Summary      Get enrolled Moodle courses
// @Description  Returns the authenticated user's enrolled courses from their linked Moodle account
// @Tags         moodle
// @Produce      json
// @Success      200  {array}   moodle.Course
// @Failure      401  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Failure      500  {object}  map[string]string
// @Failure      502  {object}  map[string]string
// @Security     BearerAuth
// @Router       /integrations/moodle/courses [get]
func (c *Controller) ListCourses(ctx *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	courses, err := c.service.GetCourses(userID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrMoodleConnectionNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Moodle account is not connected"})
		case errors.Is(err, entity.ErrMoodleTokenKeyMissing):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Server Moodle encryption is not configured"})
		default:
			ctx.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusOK, courses)
}

// ImportBoard godoc
// @Summary      Import a Moodle course as a board
// @Description  Creates a standard board with default lists and imports supported Moodle activities into the first list
// @Tags         moodle
// @Accept       json
// @Produce      json
// @Param        body  body      importBoardRequest   true  "Course import payload"
// @Success      201   {object}  importBoardResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Failure      502   {object}  map[string]string
// @Security     BearerAuth
// @Router       /integrations/moodle/import-board [post]
func (c *Controller) ImportBoard(ctx *gin.Context) {
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var request importBoardRequest
	if err := ctx.BindJSON(&request); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	boardID, err := c.service.ImportCourseAsBoard(context.Background(), userID, request.CourseID)
	if err != nil {
		switch {
		case errors.Is(err, entity.ErrMoodleConnectionNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Moodle account is not connected"})
		case errors.Is(err, entity.ErrMoodleCourseNotFound):
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Moodle course not found for this user"})
		case errors.Is(err, entity.ErrMoodleTokenKeyMissing):
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Server Moodle encryption is not configured"})
		default:
			ctx.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		}
		return
	}

	ctx.JSON(http.StatusCreated, importBoardResponse{BoardID: boardID})
}
