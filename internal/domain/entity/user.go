package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type User struct {
	BaseModel

	Email          string `gorm:"uniqueIndex;not null"`
	Name           string `gorm:"type:varchar(500)"`
	HashedPassword string `gorm:"not null"`

	Boards []Board `gorm:"foreignKey:OwnerID"`
}

type RefreshToken struct {
	BaseModel

	TokenHash string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	UserID    string    `gorm:"type:uuid;not null;index"`
	ExpiresAt time.Time `gorm:"not null"`
}

func (r *RefreshToken) HashToken(token string) {
	hash := sha256.Sum256([]byte(token))
	r.TokenHash = hex.EncodeToString(hash[:])
}

func (r *RefreshToken) IsValid() bool {
	return time.Now().Before(r.ExpiresAt)
}
