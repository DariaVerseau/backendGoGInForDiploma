package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DariaVerseau/backendGoGInForDiploma/internal/auth"
	"github.com/DariaVerseau/backendGoGInForDiploma/internal/handlers"
	"github.com/DariaVerseau/backendGoGInForDiploma/internal/health"
	"github.com/DariaVerseau/backendGoGInForDiploma/internal/middleware"
	"github.com/DariaVerseau/backendGoGInForDiploma/internal/mlclient"
	"github.com/DariaVerseau/backendGoGInForDiploma/internal/repositories"
	"github.com/DariaVerseau/backendGoGInForDiploma/internal/services"
	"github.com/DariaVerseau/backendGoGInForDiploma/internal/storage"
	"github.com/joho/godotenv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, using environment variables")
	}

	// === 1. Конфигурация ===
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET is not set")
	}
	tokenManager, err := auth.NewTokenManager(secret)
	if err != nil {
		log.Fatal("invalid JWT config: ", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	// === 2. Подключение к БД ===
	config, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatal("parse db url: ", err)
	}
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	defer pool.Close()

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer pingCancel()
	if err := pool.Ping(pingCtx); err != nil {
		log.Fatal("Database ping failed: ", err)
	}
	log.Println("Connected to PostgreSQL")

	// === 3. Инициализация зависимостей ===
	uploadDir := os.Getenv("UPLOAD_DIR")
	if uploadDir == "" {
		uploadDir = "./uploads"
	}
	fileStorage := storage.NewLocalStorage(uploadDir)

	mlClient := mlclient.NewClient()

	userRepo := repositories.NewUserRepository(pool)
	imageRepo := repositories.NewImageRepository(pool)

	authService := services.NewAuthService(userRepo, tokenManager)
	imageService := services.NewImageService(imageRepo, fileStorage, mlClient)

	authHandler := handlers.NewAuthHandler(authService)
	imageHandler := handlers.NewImageHandler(imageService)
	mlHandler := handlers.NewMLHandler(imageService)

	// === 4. Gin-роутер ===
	router := gin.Default()

	if err := router.SetTrustedProxies([]string{
		"127.0.0.1",
		"::1",
		"172.16.0.0/12",
	}); err != nil {
		log.Fatal("failed to set trusted proxies: ", err)
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.Static("/uploads", uploadDir)

	// === 5. Health-checks ===
	healthHandler := health.NewHandler(map[string]health.Pinger{
		"database":   pool,
		"ml_service": mlClient,
	})
	router.GET("/health", healthHandler.Live)
	router.GET("/ready", healthHandler.Ready)

	// === 6. API роуты ===
	v1 := router.Group("/api/v1")
	v1.POST("/register", authHandler.Register)
	v1.POST("/login", authHandler.Login)

	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(tokenManager))
	{
		protected.POST("/change-password", authHandler.ChangePassword)
		protected.GET("/images/:id", imageHandler.GetImage)
		protected.GET("/images", imageHandler.GetImages)
		protected.POST("/images", imageHandler.UploadImage)

		protected.POST("/ml/upscale", mlHandler.Upscale)
		// protected.POST("/ml/process", mlHandler.Process)
		protected.POST("/ml/enhance", mlHandler.Enhance)
		protected.POST("/ml/postprocess", mlHandler.PostProcess)
		protected.POST("/ml/style_transfer", mlHandler.StyleTransfer)
	}

	// === 7. HTTP-сервер с таймаутами ===
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadTimeout:       15 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// === 8. Graceful shutdown ===
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server stopped")
}
