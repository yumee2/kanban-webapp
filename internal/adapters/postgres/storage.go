package storage

import (
	"errors"
	"fmt"
	"student-kanban/internal/domain/entity"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type storage struct {
	db *gorm.DB
}

func NewPostgresStorage(db *gorm.DB) *storage {
	return &storage{db: db}
}

func (s *storage) CreateNewUser(user *entity.User) (string, error) {
	const fn = "adapters.repository.CreateNewUser"

	result := s.db.Create(&user)
	var pgErr *pgconn.PgError

	if result.Error != nil {
		if errors.As(result.Error, &pgErr) && pgErr.Code == "23505" {
			return "", entity.ErrEmailExist
		}
		return "", fmt.Errorf("%s: %w", fn, result.Error)
	}

	return user.ID.String(), nil
}

func (s *storage) GetUserByEmail(email string) (*entity.User, error) {
	const fn = "adapters.repository.GetUserByEmail"
	var user *entity.User

	result := s.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return &entity.User{}, entity.ErrEmailNotFound
		}

		return &entity.User{}, fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	return user, nil
}

func (s *storage) CreateRefreshToken(token *entity.RefreshToken) error {
	const fn = "adapters.repository.CreateRefreshToken"

	result := s.db.Create(&token)
	if result.Error != nil {
		return fmt.Errorf("%s: %w", fn, result.Error)
	}
	return nil
}
func (s *storage) ValidateRefreshToken(tokenValue string) (*entity.RefreshToken, error) {
	const fn = "adapters.repository.ValidateRefreshToken"

	var token entity.RefreshToken

	result := s.db.Where("token_hash = ?", tokenValue).First(&token)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return &entity.RefreshToken{}, entity.ErrRefreshTokenNotFound
		}

		return &entity.RefreshToken{}, fmt.Errorf("%s: database error: %w", fn, result.Error)
	}

	return &token, nil
}
