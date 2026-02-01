package main

import (
	"fmt"
	"student-kanban/internal/config"
	"student-kanban/internal/domain/entity"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	cfg := config.MustLoad()

	_, err := MustInitDB(cfg)
	if err != nil {
		panic(err)
	}
}

func MustInitDB(cfg *config.Config) (*gorm.DB, error) {

	dsn := fmt.Sprintf("host=%s user=%s "+
		"password=%s dbname=%s port=%d sslmode=disable",
		cfg.PostgresConnect.Host, cfg.PostgresConnect.User, cfg.PostgresConnect.Password, cfg.PostgresConnect.DatabaseName, cfg.PostgresConnect.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&entity.User{}, &entity.Board{}, &entity.List{}, &entity.Card{})

	if err != nil {
		return nil, err
	}

	return db, nil
}
