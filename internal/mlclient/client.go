//  mlclient предоставляет http-клиент для взаимодействия с ml-микросервисом
package mlclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultBaseURL        = "http://ml-service:8000"
	defaultRequestTimeout = 600 * time.Second
	pingTimeout           = 3 * time.Second
)

type Client struct {
	baseURL string
	client  *http.Client
}

func NewClient() *Client {
	baseURL := os.Getenv("ML_SERVICE_URL")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	return &Client{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: defaultRequestTimeout,
		},
	}
}

func (c *Client) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("build ping request: %w", err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("ping ml-service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ml-service unhealthy: %s", resp.Status)
	}
	return nil
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

	// НЕ передаём токен — ML сервис не требует авторизации.
	// ML работает во внутренней сети Docker.

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

	return c.doRequest(ctx, http.MethodPost, c.baseURL+endpoint, body, writer.FormDataContentType())
}

func (c *Client) PostFile(ctx context.Context, endpoint, paramName, filename string, data []byte) ([]byte, error) {
	return c.PostFileWithFields(ctx, endpoint, paramName, filename, data, nil)
}

/*func (c *Client) BasicStyleTransfer(
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
	return c.doRequest(ctx, http.MethodPost, c.baseURL+"/process", body, writer.FormDataContentType())
}
*/

func (c *Client) StyleTransfer(
	ctx context.Context,
	contentData []byte,
	contentName, styleName string,
	alpha float32,
	preserveColor bool,
) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("image", filepath.Base(contentName))
	if err != nil {
		return nil, fmt.Errorf("ошибка создания content-формы: %w", err)
	}
	if _, err := part.Write(contentData); err != nil {
		return nil, fmt.Errorf("ошибка записи content: %w", err)
	}

	if err := writer.WriteField("style", styleName); err != nil {
		return nil, fmt.Errorf("ошибка записи поля style: %w", err)
	}
	if err := writer.WriteField("alpha", fmt.Sprintf("%.1f", alpha)); err != nil {
		return nil, fmt.Errorf("ошибка записи поля alpha: %w", err)
	}
	if preserveColor {
		if err := writer.WriteField("preserve_color", "true"); err != nil {
			return nil, fmt.Errorf("ошибка записи поля preserve_color: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("ошибка закрытия формы style_transfer: %w", err)
	}

	return c.doRequest(ctx, http.MethodPost, c.baseURL+"/style_transfer_adain", body, writer.FormDataContentType())
}
