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

	"github.com/google/uuid"
)

// ImageService управляет жизненным циклом изображений
type ImageService struct {
	imageRepo *repositories.ImageRepository
	storage   *storage.LocalStorage
}

func NewImageService(imageRepo *repositories.ImageRepository, storage *storage.LocalStorage) *ImageService {
	return &ImageService{
		imageRepo: imageRepo,
		storage:   storage,
	}
}

// GetImagesByUser возвращает список изображений для указанного пользователя
func (s *ImageService) GetImagesByUser(ctx context.Context, userID int) ([]models.Image, error) {
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
		return nil, fmt.Errorf("ошибка получения изображения: %w", err)
	}

	if img == nil {
		return nil, errors.New("изображение не найдено")
	}

	return img, nil
}

// UploadImage обрабатывает загрузку файла и сохранение метаданных
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

	// Ограничение размера
	if file.Size > 10<<20 { // 10 МБ
		return nil, errors.New("файл слишком большой (макс. 10 МБ)")
	}

	// Проверка типа файла
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

	// Генерируем один UUID для ID и имени файла
	newID := uuid.New().String()

	// Создаём запись в БД
	img := &models.Image{
		ID:     newID,
		UserID: userID,
		Title:  title,
		URL:    url,
		Style:  style,
	}

	if err := s.imageRepo.Create(ctx, img); err != nil {
		return nil, fmt.Errorf("не удалось сохранить метаданные: %w", err)
	}

	// Возвращаем данные из БД (с корректным created_at)
	return s.imageRepo.GetByIDAndUser(ctx, newID, userID)
}
