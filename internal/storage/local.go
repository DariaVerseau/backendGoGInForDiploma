package storage

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type LocalStorage struct {
	uploadDir string
}

func NewLocalStorage(uploadDir string) *LocalStorage {
	os.MkdirAll(uploadDir, os.ModePerm)
	return &LocalStorage{uploadDir: uploadDir}
}

// Save сохраняет файл и возвращает относительный путь
func (s *LocalStorage) Save(file *multipart.FileHeader) (string, error) {
	// Генерируем уникальное имя: UUID + расширение
	ext := filepath.Ext(file.Filename)
	filename := uuid.New().String() + ext

	dstPath := filepath.Join(s.uploadDir, filename)
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer src.Close()

	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("не удалось создать файл: %w", err)
	}
	defer dst.Close()

	if _, err = dst.ReadFrom(src); err != nil {
		return "", fmt.Errorf("не удалось записать файл: %w", err)
	}

	// Возвращаем URL относительно корня сервера
	return "/uploads/" + filename, nil
}
