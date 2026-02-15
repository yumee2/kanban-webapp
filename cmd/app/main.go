package main

import (
	"fmt"
	"log"
	"log/slog"
	boardhttp "student-kanban/internal/adapters/http_server/boards"
	userhttp "student-kanban/internal/adapters/http_server/users"
	storage "student-kanban/internal/adapters/postgres"
	"student-kanban/internal/config"
	"student-kanban/internal/domain/entity"
	"student-kanban/internal/domain/services"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	LoadEnv()
	cfg := config.MustLoad()

	db, err := MustInitDB(cfg)
	if err != nil {
		panic(err)
	}

	userRepo := storage.NewUserStorage(db)
	userService := services.NewUserService(userRepo)

	boardRepo := storage.NewBoardStorage(db)
	boardService := services.NewBoardService(boardRepo)

	r := setUpHttpServer(userService, boardService)
	if err := r.Run(cfg.Address); err != nil {
		slog.Error("Failed to start server:", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
	}

}

func setUpHttpServer(userService userhttp.UserService, boardService boardhttp.BoardService) *gin.Engine {
	r := gin.Default()

	authController := userhttp.NewAuthController(userService)
	boardController := boardhttp.NewBoardController(boardService)

	userhttp.SetupAuthRoutes(r, authController)
	boardhttp.SetupBoardRoutes(r, boardController)
	return r
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

	err = db.AutoMigrate(&entity.User{}, &entity.Board{}, &entity.List{}, &entity.Card{}, &entity.RefreshToken{})

	if err != nil {
		return nil, err
	}

	return db, nil
}

func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}
