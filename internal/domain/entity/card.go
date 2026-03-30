package entity

import (
	"time"

	"github.com/google/uuid"
)

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
	DueDate     *time.Time

	Priority CardPriority `gorm:"type:text;default:'medium'"`

	Tags []Tag `gorm:"many2many:card_tags;"`

	Position   float64 `gorm:"not null"`
	IsFavorite bool    `gorm:"default:false"`
	IsArchived bool    `gorm:"default:false"`
}

func (p CardPriority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	}
	return false
}
