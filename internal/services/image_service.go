package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"moduleExample/web-service-gin/internal/mlclient"
	"moduleExample/web-service-gin/internal/models"
	"moduleExample/web-service-gin/internal/repositories"
	"moduleExample/web-service-gin/internal/storage"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

type ImageService struct {
	imageRepo *repositories.ImageRepository
	storage   *storage.LocalStorage
	mlClient  *mlclient.Client
}

func NewImageService(
	imageRepo *repositories.ImageRepository,
	storage *storage.LocalStorage,
	mlClient *mlclient.Client,
) *ImageService {
	return &ImageService{
		imageRepo: imageRepo,
		storage:   storage,
		mlClient:  mlClient,
	}
}

// === ВСПОМОГАТЕЛЬНЫЕ МЕТОДЫ ===

func (s *ImageService) GetImagesByUser(ctx context.Context, userID int) ([]models.Image, error) {
	images, err := s.imageRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить изображения: %w", err)
	}
	return images, nil
}

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

func (s *ImageService) UploadImage(ctx context.Context, userID int, file *multipart.FileHeader, title, style string) (*models.Image, error) {
	if file == nil || file.Size == 0 {
		return nil, errors.New("файл не предоставлен")
	}
	if title == "" {
		return nil, errors.New("название обязательно")
	}

	if style == "" {
		style = "original"
	}

	if file.Size > 10<<20 {
		return nil, errors.New("файл слишком большой (макс. 10 МБ)")
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedTypes := map[string]bool{".jpg": true, ".jpeg": true, ".png": true}
	if !allowedTypes[ext] {
		return nil, errors.New("недопустимый тип файла (разрешены: jpg, jpeg, png)")
	}

	url, err := s.storage.Save(file)
	if err != nil {
		return nil, fmt.Errorf("не удалось сохранить файл: %w", err)
	}

	newID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации UUID: %w", err)
	}

	img := &models.Image{
		ID:     newID.String(),
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

func (s *ImageService) SaveMLResult(ctx context.Context, userID int, imageData []byte, originalFilename, title, style string) (*models.Image, error) {
	if len(imageData) == 0 {
		return nil, errors.New("данные изображения пусты")
	}
	if title == "" {
		return nil, errors.New("название обязательно")
	}
	if style == "" {
		return nil, errors.New("стиль обязателен")
	}

	ext := filepath.Ext(originalFilename)
	if ext == "" {
		ext = ".png"
	}

	url, err := s.storage.SaveBytes(imageData, "ml_"+originalFilename)
	if err != nil {
		return nil, fmt.Errorf("не удалось сохранить ML-результат: %w", err)
	}

	newID, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("ошибка генерации UUID: %w", err)
	}

	img := &models.Image{
		ID:     newID.String(),
		UserID: userID,
		Title:  title,
		URL:    url,
		Style:  style,
	}

	if err := s.imageRepo.Create(ctx, img); err != nil {
		return nil, fmt.Errorf("не удалось сохранить метаданные ML-результата: %w", err)
	}
	return img, nil
}

// === ОСНОВНАЯ УНИВЕРСАЛЬНАЯ ФУНКЦИЯ ===

type mlOperation struct {
	endpoint   string
	title      string
	styleTag   string
	needsStyle bool
	params     map[string]string
}

func (s *ImageService) processML(
	ctx context.Context,
	userID int,
	contentFile *multipart.FileHeader,
	styleName string,
	op mlOperation,
) (*models.Image, error) {
	// ❌ Убираем установку токена - не нужно
	// token := getTokenFromContext(ctx)
	// s.mlClient.SetAuthToken(token)

	contentData, err := readFile(contentFile)
	if err != nil {
		return nil, err
	}

	var resultData []byte

	if op.endpoint == "/style_transfer_adain" {
		resultData, err = s.mlClient.StyleTransfer(ctx, contentData, contentFile.Filename, styleName)
	} else if len(op.params) > 0 {
		resultData, err = s.mlClient.PostFileWithFields(ctx, op.endpoint, "image", contentFile.Filename, contentData, op.params)
	} else {
		resultData, err = s.mlClient.PostFile(ctx, op.endpoint, "image", contentFile.Filename, contentData)
	}

	if err != nil {
		return nil, fmt.Errorf("ошибка ML-обработки %s: %w", op.endpoint, err)
	}

	if len(resultData) == 0 {
		return nil, fmt.Errorf("ML-сервис %s вернул пустой ответ", op.endpoint)
	}

	return s.SaveMLResult(ctx, userID, resultData, contentFile.Filename, op.title, op.styleTag)
}

func readFile(file *multipart.FileHeader) ([]byte, error) {
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer src.Close()

	data, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать файл: %w", err)
	}
	return data, nil
}

// === ПУБЛИЧНЫЕ МЕТОДЫ ===

func (s *ImageService) Upscale(ctx context.Context, userID int, file *multipart.FileHeader, scale int) (*models.Image, error) {
	if scale != 2 && scale != 4 {
		return nil, errors.New("scale must be 2 or 4")
	}
	return s.processML(ctx, userID, file, "", mlOperation{
		endpoint: "/upscale",
		title:    fmt.Sprintf("Upscaled x%d", scale),
		styleTag: fmt.Sprintf("upscale_x%d", scale),
		params:   map[string]string{"scale": strconv.Itoa(scale)},
	})
}

func (s *ImageService) Enhance(ctx context.Context, userID int, file *multipart.FileHeader, fidelityWeight float64, postprocess bool) (*models.Image, error) {
	return s.processML(ctx, userID, file, "", mlOperation{
		endpoint: "/enhance",
		title:    "Enhanced Image",
		styleTag: "enhance",
		params: map[string]string{
			"fidelity_weight": strconv.FormatFloat(fidelityWeight, 'f', 2, 64),
			"postprocess":     strconv.FormatBool(postprocess),
		},
	})
}

func (s *ImageService) PostProcess(ctx context.Context, userID int, file *multipart.FileHeader) (*models.Image, error) {
	return s.processML(ctx, userID, file, "", mlOperation{
		endpoint: "/postprocess",
		title:    "Postprocessed Image",
		styleTag: "postprocess",
	})
}

func (s *ImageService) StyleTransfer(
	ctx context.Context,
	userID int,
	contentFile *multipart.FileHeader,
	styleName string,
) (*models.Image, error) {
	// ❌ Убираем установку токена
	// token := getTokenFromContext(ctx)
	// s.mlClient.SetAuthToken(token)

	contentData, err := readFile(contentFile)
	if err != nil {
		return nil, err
	}

	resultData, err := s.mlClient.StyleTransfer(ctx, contentData, contentFile.Filename, styleName)
	if err != nil {
		return nil, fmt.Errorf("ошибка ML-style_transfer: %w", err)
	}

	return s.SaveMLResult(ctx, userID, resultData, contentFile.Filename, "Styled Image", "style_transfer")
}

func (s *ImageService) BasicStyleTransfer(
	ctx context.Context,
	userID int,
	contentFile *multipart.FileHeader,
	styleName string,
) (*models.Image, error) {
	// ❌ Убираем установку токена
	// token := getTokenFromContext(ctx)
	// s.mlClient.SetAuthToken(token)

	supportedStyles := map[string]bool{
		"vangogh":    true,
		"picasso":    true,
		"monet":      true,
		"monet2":     true,
		"erinHanson": true,
		"sketch":     true,
	}
	if !supportedStyles[styleName] {
		return nil, fmt.Errorf("неподдерживаемый стиль: %s", styleName)
	}

	contentData, err := readFile(contentFile)
	if err != nil {
		return nil, err
	}

	resultData, err := s.mlClient.BasicStyleTransfer(ctx, contentData, contentFile.Filename, styleName)
	if err != nil {
		return nil, fmt.Errorf("ошибка базового NST: %w", err)
	}

	return s.SaveMLResult(ctx, userID, resultData, contentFile.Filename, "Basic Styled Image", "basic_nst")
}

func (s *ImageService) PostProcessWithParams(ctx context.Context, userID int, file *multipart.FileHeader, params map[string]string) (*models.Image, error) {
	// ❌ Убираем установку токена
	// token := getTokenFromContext(ctx)
	// s.mlClient.SetAuthToken(token)

	return s.processML(ctx, userID, file, "", mlOperation{
		endpoint: "/postprocess",
		title:    "Postprocessed Image",
		styleTag: "postprocess",
		params:   params,
	})
}

func (s *ImageService) Colorize(ctx context.Context, userID int, file *multipart.FileHeader) (*models.Image, error) {
	// ❌ Убираем установку токена
	// token := getTokenFromContext(ctx)
	// s.mlClient.SetAuthToken(token)

	if file.Size > 10<<20 {
		return nil, errors.New("файл слишком большой (макс. 10 МБ)")
	}

	contentData, err := readFile(file)
	if err != nil {
		return nil, err
	}

	resultData, err := s.mlClient.PostFile(ctx, "/colorize", "image", file.Filename, contentData)
	if err != nil {
		return nil, fmt.Errorf("ошибка colorize: %w", err)
	}

	return s.SaveMLResult(ctx, userID, resultData, file.Filename, "Colorized Image", "colorize")
}

func (s *ImageService) Process(
	ctx context.Context,
	userID int,
	contentFile *multipart.FileHeader,
	styleName string,
) (*models.Image, error) {
	return s.BasicStyleTransfer(ctx, userID, contentFile, styleName)
}
