// Package mlclient предоставляет HTTP-клиент для взаимодействия с ML-микросервисом.
package mlclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"time"
)

const MLServiceURL = "http://ml-service:8000"

type Client struct {
	client            *http.Client
	longTimeoutClient *http.Client
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{
			Timeout: 1000 * time.Second,
		},
		longTimeoutClient: &http.Client{
			Timeout: 600 * time.Second, // 10 минут для базового NST
		},
	}
}

// doRequest выполняет HTTP-запрос и возвращает тело ответа или ошибку.
func (c *Client) doRequest(ctx context.Context, method, url string, body io.Reader, contentType string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML-сервис вернул статус %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// PostFileWithFields отправляет один файл и дополнительные поля формы (например, scale).
func (c *Client) PostFileWithFields(
	ctx context.Context,
	endpoint, fileParamName, filename string,
	fileData []byte,
	fields map[string]string,
) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fileParamName, filepath.Base(filename))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания файла в форме: %w", err)
	}
	if _, err := part.Write(fileData); err != nil {
		return nil, fmt.Errorf("ошибка записи данных файла: %w", err)
	}

	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return nil, fmt.Errorf("ошибка записи поля %s: %w", key, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("ошибка закрытия формы: %w", err)
	}

	return c.doRequest(ctx, "POST", MLServiceURL+endpoint, body, writer.FormDataContentType())
}

// PostFile — отправка файла без дополнительных полей.
func (c *Client) PostFile(ctx context.Context, endpoint, paramName, filename string, data []byte) ([]byte, error) {
	return c.PostFileWithFields(ctx, endpoint, paramName, filename, data, nil)
}

// BasicStyleTransfer применяет базовый перенос стиля через эндпоинт /process.
// Использует тот же формат запроса, что и StyleTransfer, но другой URL.
func (c *Client) BasicStyleTransfer(
	ctx context.Context,
	contentData []byte,
	contentName, styleName string,
) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Content как файл
	if part, err := writer.CreateFormFile("image", filepath.Base(contentName)); err != nil {
		return nil, fmt.Errorf("ошибка создания content-формы: %w", err)
	} else if _, err := part.Write(contentData); err != nil {
		return nil, fmt.Errorf("ошибка записи content: %w", err)
	}

	// Style как строковое поле
	if err := writer.WriteField("style", styleName); err != nil {
		return nil, fmt.Errorf("ошибка записи поля style: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("ошибка закрытия формы basic_style_transfer: %w", err)
	}

	// Отправляем на /process, а не на /style_transfer_adain
	return c.doRequest(ctx, "POST", MLServiceURL+"/process", body, writer.FormDataContentType())
}

// StyleTransfer отправляет два файла: content и style.
func (c *Client) StyleTransfer(
	ctx context.Context,
	contentData []byte,
	contentName, styleName string,
) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Content как файл
	if part, err := writer.CreateFormFile("image", filepath.Base(contentName)); err != nil {
		return nil, fmt.Errorf("ошибка создания content-формы: %w", err)
	} else if _, err := part.Write(contentData); err != nil {
		return nil, fmt.Errorf("ошибка записи content: %w", err)
	}

	// Style как строковое поле
	if err := writer.WriteField("style", styleName); err != nil {
		return nil, fmt.Errorf("ошибка записи поля style: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("ошибка закрытия формы style_transfer: %w", err)
	}

	return c.doRequest(ctx, "POST", MLServiceURL+"/style_transfer_adain", body, writer.FormDataContentType())
}
