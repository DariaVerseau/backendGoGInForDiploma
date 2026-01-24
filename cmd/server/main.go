package main

import (
	"context"
	"log"
	"moduleExample/web-service-gin/internal/handlers"
	"moduleExample/web-service-gin/internal/middleware"
	"moduleExample/web-service-gin/internal/repositories"
	"moduleExample/web-service-gin/internal/services"
	"moduleExample/web-service-gin/internal/storage"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload" // автоматически загружает .env

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// === 1. Подключение к базе данных ===
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Простая проверка подключения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.Ping(ctx); err != nil {
		log.Fatal("Database ping failed:", err)
	}
	log.Println("Connected to PostgreSQL")

	uploadDir := "./uploads"
	fileStorage := storage.NewLocalStorage(uploadDir)

	// === 2. Инициализация репозиториев ===
	userRepo := repositories.NewUserRepository(db)
	imageRepo := repositories.NewImageRepository(db)

	// === 3. Инициализация сервисов ===
	authService := services.NewAuthService(userRepo)
	imageService := services.NewImageService(imageRepo, fileStorage)

	// === 4. Инициализация хендлеров ===
	authHandler := handlers.NewAuthHandler(authService)
	imageHandler := handlers.NewImageHandler(imageService)

	// === 5. Настройка Gin-роутера ===
	router := gin.Default()

	// Публичные роуты
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	// Защищённые роуты (требуют JWT)
	protected := router.Group("/")
	protected.Use(middleware.AuthMiddleware())
	{
		protected.GET("/images", imageHandler.GetImages)
		protected.POST("/images", imageHandler.UploadImage)
	}

	// === 6. Запуск сервера ===
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
