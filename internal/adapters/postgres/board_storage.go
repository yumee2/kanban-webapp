package storage

import "gorm.io/gorm"

type boardStorage struct {
	db *gorm.DB
}

func NewBoardStorage(db *gorm.DB) *boardStorage {
	return &boardStorage{db: db}
}
