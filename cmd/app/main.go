package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	boardhttp "student-kanban/internal/adapters/http_server/boards"
	cardshttp "student-kanban/internal/adapters/http_server/cards"
	listshttp "student-kanban/internal/adapters/http_server/lists"
	moodlehttp "student-kanban/internal/adapters/http_server/moodle"
	tagshttp "student-kanban/internal/adapters/http_server/tags"
	userhttp "student-kanban/internal/adapters/http_server/users"
	"student-kanban/internal/adapters/moodle"
	storage "student-kanban/internal/adapters/postgres"
	"student-kanban/internal/config"
	"student-kanban/internal/domain/entity"
	"student-kanban/internal/domain/services"

	_ "student-kanban/docs"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// @title           Student Kanban API
// @version         1.0
// @description     API for managing kanban boards
// @host            localhost:8003
// @BasePath        /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

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

	listRepo := storage.NewListRepository(db)
	listService := services.NewListService(listRepo, boardRepo)

	cardRepo := storage.NewCardRepository(db)
	cardService := services.NewCardService(cardRepo, listRepo)

	tagsRepo := storage.NewTagRepository(db)
	tagService := services.NewTagService(tagsRepo, boardRepo, cardRepo)

	moodleConnectionRepo := storage.NewMoodleConnectionStorage(db)
	moodleClient := moodle.NewClient()
	moodleConnectionService := services.NewMoodleConnectionService(moodleConnectionRepo, moodleClient, boardRepo, listRepo, cardRepo)

	r := setUpHttpServer(userService, boardService, listService, cardService, tagService, moodleConnectionService)
	if err := r.Run(cfg.Address); err != nil {
		slog.Error("Failed to start server:", slog.Attr{Key: "error", Value: slog.StringValue(err.Error())})
	}

}

func setUpHttpServer(userService userhttp.UserService, boardService boardhttp.BoardService,
	listService listshttp.ListService, cardService cardshttp.CardService, tagService tagshttp.TagService,
	moodleService moodlehttp.MoodleService) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())

	authController := userhttp.NewAuthController(userService)
	boardController := boardhttp.NewBoardController(boardService)
	listController := listshttp.NewListController(listService)
	cardController := cardshttp.NewCardController(cardService)
	tagController := tagshttp.NewTagController(tagService)
	moodleController := moodlehttp.NewController(moodleService)

	userhttp.SetupAuthRoutes(r, authController)
	boardhttp.SetupBoardRoutes(r, boardController)
	listshttp.SetupListRoutes(r, listController)
	cardshttp.SetupCardRoutes(r, cardController)
	tagshttp.SetupTagRoutes(r, tagController)
	moodlehttp.SetupRoutes(r, moodleController)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}

func corsMiddleware() gin.HandlerFunc {
	allowedOrigin := os.Getenv("CLIENT_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:5173"
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin == allowedOrigin {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
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

	err = db.AutoMigrate(&entity.User{}, &entity.Board{}, &entity.List{}, &entity.Tag{}, &entity.Card{}, &entity.RefreshToken{}, &entity.MoodleConnection{})

	if err != nil {
		return nil, err
	}

	return db, nil
}

func LoadEnv() {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Fatal("Error loading .env file")
	}
}
