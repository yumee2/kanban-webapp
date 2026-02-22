package userhttp

import (
	"errors"
	"log/slog"
	"net/http"
	"student-kanban/internal/domain/entity"
	"student-kanban/internal/domain/services"

	"github.com/gin-gonic/gin"
)

type UserService interface {
	RegisterNewUser(email, password, name string) (*services.JWTTokenPair, error)
	LoginExistingUser(email string, password string) (*services.JWTTokenPair, error)
	ValidateRefreshToken(refreshToken string) (string, error)
}

type authController struct {
	authService UserService
}

func NewAuthController(authServ UserService) *authController {
	return &authController{authService: authServ}
}

type JWTTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// Register godoc
// @Summary      Register a new user
// @Description  Creates a new user account and returns a JWT token pair
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      registerRequest   true  "Registration data"
// @Success      201   {object}  JWTTokenResponse
// @Failure      400   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /auth/register [post]
func (c *authController) Register(ctx *gin.Context) {
	const fn = "adapters.controller.Register"
	log := slog.With(
		slog.String("fn", fn),
	)

	var request registerRequest
	if err := ctx.BindJSON(&request); err != nil {
		log.Error("failed to parse json body", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jwtTokenPair, err := c.authService.RegisterNewUser(request.Email, request.Password, request.Name)
	if err != nil {
		if errors.Is(err, entity.ErrEmailExist) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "User with provided email already exists"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, JWTTokenResponse{
		AccessToken:  jwtTokenPair.AccessToken,
		RefreshToken: jwtTokenPair.RefreshToken,
	})
}

// Login godoc
// @Summary      Login user
// @Description  Authenticates a user and returns a JWT token pair
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      loginRequest      true  "Login credentials"
// @Success      201   {object}  JWTTokenResponse
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /auth/login [post]
func (c *authController) Login(ctx *gin.Context) {
	const fn = "adapters.controller.Login"
	log := slog.With(
		slog.String("fn", fn),
	)

	var request loginRequest
	if err := ctx.BindJSON(&request); err != nil {
		log.Error("failed to parse json body", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jwtTokenPair, err := c.authService.LoginExistingUser(request.Email, request.Password)
	if err != nil {
		if errors.Is(err, entity.ErrEmailNotFound) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User with provided email was not found"})
			return
		}

		if errors.Is(err, entity.ErrInvalidPassword) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, JWTTokenResponse{
		AccessToken:  jwtTokenPair.AccessToken,
		RefreshToken: jwtTokenPair.RefreshToken,
	})
}

// Refresh godoc
// @Summary      Refresh access token
// @Description  Validates a refresh token and returns a new access token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body      refreshRequest    true  "Refresh token"
// @Success      200   {object}  map[string]string
// @Failure      400   {object}  map[string]string
// @Failure      401   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Router       /auth/refresh [post]
func (c *authController) Refresh(ctx *gin.Context) {
	const fn = "adapters.controller.Refresh"
	log := slog.With(
		slog.String("fn", fn),
	)

	var request refreshRequest
	if err := ctx.BindJSON(&request); err != nil {
		log.Error("failed to parse json body", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, err := c.authService.ValidateRefreshToken(request.RefreshToken)
	if err != nil {
		if errors.Is(err, entity.ErrRefreshTokenNotFound) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Refresh token not found"})
			return
		}
		if errors.Is(err, entity.ErrRefreshTokenExpired) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired"})
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}
