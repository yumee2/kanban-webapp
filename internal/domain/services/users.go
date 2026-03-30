package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"os"
	"student-kanban/internal/domain/entity"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"golang.org/x/crypto/bcrypt"
)

type JWTTokenPair struct {
	RefreshToken       string
	AccessToken        string
	AccessExpireTime   time.Time
	RefreshExprireTime time.Time
}

type UserRepository interface {
	CreateNewUser(user *entity.User) (string, error)
	GetUserByEmail(email string) (*entity.User, error)
	CreateRefreshToken(token *entity.RefreshToken) error
	ValidateRefreshToken(tokenValue string) (*entity.RefreshToken, error)
	DeleteRefreshToken(tokenValue string) error
}

type userService struct {
	userRepo UserRepository
}

func NewUserService(userRepo UserRepository) *userService {
	return &userService{
		userRepo: userRepo,
	}
}

// Return generated access and refresh tokens or error
func (s *userService) RegisterNewUser(email, password, name string) (*JWTTokenPair, error) {
	const fn = "domain.service.RegisterNewUser"
	log := slog.With(
		slog.String("fn", fn),
	)

	hashedPassword, err := hashPassword(password)
	if err != nil {
		log.Error("failed to hash password", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return &JWTTokenPair{}, err
	}

	newUser := &entity.User{Email: email, HashedPassword: hashedPassword, Name: name}

	uuid, err := s.userRepo.CreateNewUser(newUser)
	if err != nil {
		if errors.Is(err, entity.ErrEmailExist) {
			return &JWTTokenPair{}, entity.ErrEmailExist
		}
		log.Error("failed to save new user credentials", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return &JWTTokenPair{}, err
	}

	jwtTokens, err := generateJWTTokenPair(uuid)
	if err != nil {
		log.Error("failed to create JWT tokens", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())},
			slog.Attr{Key: "UserUUID", Value: slog.StringValue(uuid)})
		return &JWTTokenPair{}, err
	}

	refreshToken := &entity.RefreshToken{TokenHash: jwtTokens.RefreshToken, UserID: uuid, ExpiresAt: jwtTokens.RefreshExprireTime}
	refreshToken.HashToken(jwtTokens.RefreshToken) // hashing the token to store in database
	if err = s.userRepo.CreateRefreshToken(refreshToken); err != nil {
		log.Error("failed to store refresh token in database", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return &JWTTokenPair{}, err
	}

	return jwtTokens, nil
}

func (s *userService) LoginExistingUser(email string, password string) (*JWTTokenPair, error) {
	const fn = "domain.service.LoginExistingUser"
	log := slog.With(
		slog.String("fn", fn),
	)

	user, err := s.userRepo.GetUserByEmail(email)
	if err != nil {
		if errors.Is(err, entity.ErrEmailNotFound) {
			log.Error("user with provided email not found", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
			return &JWTTokenPair{}, entity.ErrEmailNotFound
		}
		log.Error("failed to get user by email", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return &JWTTokenPair{}, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			log.Error("invalid password", slog.Attr{Key: "email", Value: slog.StringValue(email)})
			return &JWTTokenPair{}, entity.ErrInvalidPassword
		}
		log.Error("failed to compare password", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return &JWTTokenPair{}, err
	}

	jwtTokens, err := generateJWTTokenPair(user.ID.String())
	if err != nil {
		log.Error("failed to create JWT tokens", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())},
			slog.Attr{Key: "UserUUID", Value: slog.StringValue(user.ID.String())})
		return &JWTTokenPair{}, err
	}

	refreshToken := &entity.RefreshToken{TokenHash: jwtTokens.RefreshToken, UserID: user.ID.String(), ExpiresAt: jwtTokens.RefreshExprireTime}
	refreshToken.HashToken(jwtTokens.RefreshToken) // hashing the token to store in database
	if err = s.userRepo.CreateRefreshToken(refreshToken); err != nil {
		log.Error("failed to store refresh token in database", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return &JWTTokenPair{}, err
	}

	return jwtTokens, nil
}

// Return newly generated access token or error
func (s *userService) RefreshSession(refreshToken string) (*JWTTokenPair, error) {
	const fn = "domain.service.ValidateRefreshToken"
	log := slog.With(
		slog.String("fn", fn),
	)

	hashedToken := hashToken(refreshToken)
	refresh, err := s.userRepo.ValidateRefreshToken(hashedToken)
	if err != nil {
		if errors.Is(err, entity.ErrRefreshTokenNotFound) {
			log.Error("provided token not found", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
			return &JWTTokenPair{}, entity.ErrRefreshTokenNotFound
		}
		log.Error("failed to get a refresh token from database", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return &JWTTokenPair{}, err
	}

	if !refresh.IsValid() { // check if the token is expired or not
		return &JWTTokenPair{}, entity.ErrRefreshTokenExpired
	}

	jwtTokens, err := generateJWTTokenPair(refresh.UserID)
	if err != nil {
		log.Error("failed to create JWT tokens", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())},
			slog.Attr{Key: "UserUUID", Value: slog.StringValue(refresh.UserID)})
		return &JWTTokenPair{}, err
	}

	if err = s.userRepo.DeleteRefreshToken(hashedToken); err != nil {
		log.Error("failed to delete old refresh token", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return &JWTTokenPair{}, err
	}

	newRefreshToken := &entity.RefreshToken{TokenHash: jwtTokens.RefreshToken, UserID: refresh.UserID, ExpiresAt: jwtTokens.RefreshExprireTime}
	newRefreshToken.HashToken(jwtTokens.RefreshToken)
	if err = s.userRepo.CreateRefreshToken(newRefreshToken); err != nil {
		log.Error("failed to store refreshed token in database", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
		return &JWTTokenPair{}, err
	}

	return jwtTokens, nil
}

func (s *userService) Logout(refreshToken string) error {
	if refreshToken == "" {
		return nil
	}

	return s.userRepo.DeleteRefreshToken(hashToken(refreshToken))
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}
func generateJWTTokenPair(userUUID string) (*JWTTokenPair, error) {
	var (
		jwtSecret  = os.Getenv("JWT_SECRET")
		accessTTL  = 60 * time.Minute
		refreshTTL = 14 * 24 * time.Hour
	)

	accessExpire := time.Now().Add(accessTTL)
	refreshExpire := time.Now().Add(refreshTTL)

	accessClaims := jwt.MapClaims{
		"user_id": userUUID,
		"exp":     accessExpire.Unix(),
	}

	refreshClaims := jwt.MapClaims{
		"user_id": userUUID,
		"exp":     refreshExpire.Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return &JWTTokenPair{}, err
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return &JWTTokenPair{}, err
	}

	return &JWTTokenPair{AccessToken: accessTokenString, RefreshToken: refreshTokenString,
		RefreshExprireTime: refreshExpire, AccessExpireTime: accessExpire}, nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	hashedToken := hex.EncodeToString(hash[:])
	return hashedToken
}
