package services

import (
	"context"
	"moduleExample/web-service-gin/internal/mlclient"
	"moduleExample/web-service-gin/internal/storage"
)

type MLService struct {
	client    *mlclient.Client
	storage   *storage.LocalStorage
	uploadDir string
}

func NewMLService(storage *storage.LocalStorage, uploadDir string) *MLService {
	return &MLService{
		client:    mlclient.NewClient(),
		storage:   storage,
		uploadDir: uploadDir,
	}
}

func (s *MLService) UpscaleImage(ctx context.Context, imageData []byte, originalFilename string) (string, error) {
	result, err := s.client.Upscale(ctx, imageData, originalFilename)
	if err != nil {
		return "", err
	}

	// Используем SaveBytes вместо Save
	resultURL, err := s.storage.SaveBytes(result, originalFilename)
	if err != nil {
		return "", err
	}

	return resultURL, nil
}

func (s *MLService) EnhanceFace(ctx context.Context, imageData []byte, originalFilename string) (string, error) {
	result, err := s.client.EnhanceFace(ctx, imageData, originalFilename)
	if err != nil {
		return "", err
	}

	// Используем SaveBytes вместо Save
	resultURL, err := s.storage.SaveBytes(result, originalFilename)
	if err != nil {
		return "", err
	}

	return resultURL, nil
}

func (s *MLService) ColorizeImage(ctx context.Context, imageData []byte, originalFilename string) (string, error) {
	result, err := s.client.Colorize(ctx, imageData, originalFilename)
	if err != nil {
		return "", err
	}
	return s.storage.SaveBytes(result, "colorized_"+originalFilename)
}

func (s *MLService) ApplyStyle(ctx context.Context, contentData, styleData []byte, contentName, styleName string) (string, error) {
	result, err := s.client.StyleTransferAdaIn(ctx, contentData, styleData, contentName, styleName)
	if err != nil {
		return "", err
	}
	return s.storage.SaveBytes(result, "styled_"+contentName)
}

// добавить все эндпоинты
