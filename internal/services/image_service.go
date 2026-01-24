package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"moduleExample/web-service-gin/internal/models"
	"moduleExample/web-service-gin/internal/repositories"
	"moduleExample/web-service-gin/internal/storage"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// ImageService управляет жизненным циклом изображений
type ImageService struct {
	imageRepo *repositories.ImageRepository
	storage   *storage.LocalStorage // added
	// Позже сюда добавим: mlClient, storage и т.д.
}

func NewImageService(imageRepo *repositories.ImageRepository, storage *storage.LocalStorage) *ImageService {
	return &ImageService{
		imageRepo: imageRepo,
		storage:   storage,
	}
}

// generateImageID генерирует уникальный ID для изображения
func generateImageID() string {
	return uuid.New().String()
}

// CreateImage создаёт новое изображение с UUID
func (s *ImageService) CreateImage(ctx context.Context, userID int, title, style string) (*models.Image, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}
	if style == "" {
		return nil, errors.New("style is required")
	}

	img := &models.Image{
		ID:     generateImageID(), // ← теперь UUID!
		UserID: userID,
		Title:  title,
		Style:  style,
		URL:    "/uploads/" + generateImageID() + ".jpg", // временный URL
	}

	err := s.imageRepo.Create(ctx, img)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// GetImagesByUser возвращает список изображений для указанного пользователя
func (s *ImageService) GetImagesByUser(ctx context.Context, userID int) ([]models.Image, error) {
	// Делегируем запрос репозиторию
	images, err := s.imageRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить изображения: %w", err)
	}
	return images, nil
}

// GetImageByID проверяет, принадлежит ли изображение пользователю
func (s *ImageService) GetImageByID(ctx context.Context, imageID string, userID int) (*models.Image, error) {
	img, err := s.imageRepo.GetByIDAndUser(ctx, imageID, userID)
	if err != nil {
		return nil, err
	}
	if img == nil {
		return nil, errors.New("изображение не найдено")
	}
	return img, nil
}

// Вспомогательная функция для генерации ID
func generateID() string {
	// В реальном проекте используй UUID
	return time.Now().Format("20060102150405") // YYYYMMDDHHMMSS
}

func (s *ImageService) UploadImage(ctx context.Context, userID int, file *multipart.FileHeader, title, style string) (*models.Image, error) {
	// Валидация
	if file == nil {
		return nil, errors.New("файл не предоставлен")
	}
	if title == "" {
		return nil, errors.New("название обязательно")
	}
	if style == "" {
		return nil, errors.New("стиль обязателен")
	}

	// Ограничение размера (опционально)
	if file.Size > 10<<20 { // 10 МБ
		return nil, errors.New("файл слишком большой (макс. 10 МБ)")
	}

	// Разрешённые типы (опционально)
	allowedTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}
	ext := filepath.Ext(file.Filename)
	if !allowedTypes[ext] {
		return nil, errors.New("недопустимый тип файла (разрешены: jpg, jpeg, png)")
	}

	// Сохраняем файл
	url, err := s.storage.Save(file)
	if err != nil {
		return nil, fmt.Errorf("не удалось сохранить файл: %w", err)
	}

	// Создаём запись в БД
	img := &models.Image{
		ID:     uuid.New().String(),
		UserID: userID,
		Title:  title,
		URL:    url,
		Style:  style,
	}

	if err := s.imageRepo.Create(ctx, img); err != nil {
		return nil, fmt.Errorf("не удалось сохранить метаданные: %w", err)
	}

	return img, nil
}
