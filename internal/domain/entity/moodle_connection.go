package entity

import "github.com/google/uuid"

type MoodleConnection struct {
	BaseModel

	UserID           uuid.UUID `gorm:"type:uuid;not null;uniqueIndex"`
	User             User      `gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	BaseURL          string    `gorm:"type:text;not null"`
	MoodleUserID     int64     `gorm:"not null"`
	MoodleUsername   string    `gorm:"type:varchar(255);not null"`
	MoodleFullName   string    `gorm:"type:varchar(255);not null"`
	MoodleSiteName   string    `gorm:"type:varchar(255);not null"`
	EncryptedToken   string    `gorm:"type:text;not null"`
	TokenFingerprint string    `gorm:"type:varchar(64);not null"`
	ServiceShortName string    `gorm:"type:varchar(255);not null"`
}
