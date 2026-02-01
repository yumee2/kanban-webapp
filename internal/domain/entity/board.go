package entity

import "github.com/google/uuid"

type Board struct {
	BaseModel

	OwnerID uuid.UUID `gorm:"type:uuid;not null"`
	Owner   User      `gorm:"foreignKey:OwnerID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Title       string  `gorm:"type:varchar(100);not null"`
	Description *string `gorm:"type:varchar(500)"`

	Lists []List `gorm:"foreignKey:BoardID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}
