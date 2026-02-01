package entity

type User struct {
	BaseModel

	Email          string `gorm:"uniqueIndex;not null"`
	Name           string `gorm:"type:varchar(500)"`
	HashedPassword string `gorm:"not null"`

	Boards []Board `gorm:"foreignKey:OwnerID"`
}
