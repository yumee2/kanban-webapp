package entity

import "github.com/google/uuid"

type List struct {
	BaseModel

	BoardID uuid.UUID `gorm:"type:uuid;not null;index"`
	Board   Board     `gorm:"foreignKey:BoardID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Title    string  `gorm:"not null"`
	Position float64 `gorm:"not null"` // ordering

	Cards []Card `gorm:"foreignKey:ListID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
