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

func (s *LocalStorage) GetFullPath(relativePath string) string {
	// Если путь уже абсолютный
	if filepath.IsAbs(relativePath) {
		return relativePath
	}
	// Иначе добавляем basePath
	return filepath.Join(s.uploadDir, filepath.Base(relativePath))
}

// GetBasePath возвращает базовый путь к хранилищу
func (s *LocalStorage) GetBasePath() string {
	return s.uploadDir
}

// Save сохраняет загруженный файл из HTTP-запроса
func (s *LocalStorage) Save(file *multipart.FileHeader) (string, error) {
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

	return "uploads/" + filename, nil
}

// SaveBytes сохраняет байты как файл и возвращает URL
func (s *LocalStorage) SaveBytes(data []byte, originalFilename string) (string, error) {
	// Генерируем уникальное имя с оригинальным расширением
	ext := filepath.Ext(originalFilename)
	if ext == "" {
		ext = ".png" // по умолчанию для изображений
	}
	filename := uuid.New().String() + ext
	dstPath := filepath.Join(s.uploadDir, filename)

	if err := os.WriteFile(dstPath, data, 0644); err != nil {
		return "", fmt.Errorf("не удалось сохранить байты в файл: %w", err)
	}

	return "uploads/" + filename, nil
}
