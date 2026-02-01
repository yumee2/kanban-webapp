package entity

import "github.com/google/uuid"

type CardPriority string

const (
	PriorityLow    CardPriority = "low"
	PriorityMedium CardPriority = "medium"
	PriorityHigh   CardPriority = "high"
)

type Card struct {
	BaseModel

	ListID uuid.UUID `gorm:"type:uuid;not null;index"`
	List   List      `gorm:"foreignKey:ListID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	Title       string  `gorm:"not null"`
	Description *string `gorm:"type:varchar(500)"`

	Priority CardPriority `gorm:"type:text;default:'medium'"`

	Position   float64 `gorm:"not null"`
	IsArchived bool    `gorm:"default:false"`
}
