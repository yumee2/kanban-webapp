package entity

import "github.com/google/uuid"

type Tag struct {
	BaseModel
	BoardID uuid.UUID `gorm:"type:uuid;not null;index"`
	Board   Board     `gorm:"foreignKey:BoardID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Name    string    `gorm:"not null"`
	Color   string    `gorm:"not null"`
}
