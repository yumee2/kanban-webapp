package storage

import (
	"errors"
	"fmt"
	"student-kanban/internal/domain/entity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type moodleConnectionStorage struct {
	db *gorm.DB
}

func NewMoodleConnectionStorage(db *gorm.DB) *moodleConnectionStorage {
	return &moodleConnectionStorage{db: db}
}

func (s *moodleConnectionStorage) Upsert(connection *entity.MoodleConnection) error {
	const fn = "adapters.repository.UpsertMoodleConnection"

	var existing entity.MoodleConnection
	err := s.db.Where("user_id = ?", connection.UserID).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("%s: %w", fn, err)
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := s.db.Create(connection).Error; err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		}
		return nil
	}

	existing.BaseURL = connection.BaseURL
	existing.MoodleUserID = connection.MoodleUserID
	existing.MoodleUsername = connection.MoodleUsername
	existing.MoodleFullName = connection.MoodleFullName
	existing.MoodleSiteName = connection.MoodleSiteName
	existing.EncryptedToken = connection.EncryptedToken
	existing.TokenFingerprint = connection.TokenFingerprint
	existing.ServiceShortName = connection.ServiceShortName

	if err := s.db.Save(&existing).Error; err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func (s *moodleConnectionStorage) GetByUserID(userID uuid.UUID) (*entity.MoodleConnection, error) {
	const fn = "adapters.repository.GetMoodleConnectionByUserID"

	var connection entity.MoodleConnection
	if err := s.db.Where("user_id = ?", userID).First(&connection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, entity.ErrMoodleConnectionNotFound
		}
		return nil, fmt.Errorf("%s: %w", fn, err)
	}

	return &connection, nil
}
