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

// MLServiceURL — имя сервиса в docker-compose + порт внутри сети Docker
const MLServiceURL = "http://ml-service:8000"

type Client struct {
	client *http.Client
}

func NewClient() *Client {
	return &Client{
		client: &http.Client{
			Timeout: 90 * time.Second, // ML может обрабатывать долго
		},
	}
}

func (c *Client) postFile(ctx context.Context, endpoint, paramName, filename string, data []byte) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(paramName, filepath.Base(filename))
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	_, err = part.Write(data)
	if err != nil {
		return nil, fmt.Errorf("write file data: %w", err)
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", MLServiceURL+endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ML service error: status %d", resp.StatusCode)
	}

	result, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return result, nil
}

// Upscale отправляет изображение в /upscale
func (c *Client) Upscale(ctx context.Context, imageData []byte, filename string) ([]byte, error) {
	return c.postFile(ctx, "/upscale", "file", filename, imageData)
}

// EnhanceFace отправляет изображение в /enhance_face
func (c *Client) EnhanceFace(ctx context.Context, imageData []byte, filename string) ([]byte, error) {
	return c.postFile(ctx, "/enhance", "file", filename, imageData)
}

// StyleTransfer отправляет content и style (упрощённо — можно расширить)
func (c *Client) StyleTransfer(ctx context.Context, contentData, styleData []byte) ([]byte, error) {
	// Для двух файлов нужен отдельный метод — пока оставим как заглушку
	return nil, fmt.Errorf("style transfer not implemented in client yet")
}

// Colorize вызывает /colorize
func (c *Client) Colorize(ctx context.Context, imageData []byte, filename string) ([]byte, error) {
	return c.postFile(ctx, "/colorize", "file", filename, imageData)
}

// Postprocess вызывает /postprocess
func (c *Client) Postprocess(ctx context.Context, imageData []byte, filename string) ([]byte, error) {
	return c.postFile(ctx, "/postprocess", "file", filename, imageData)
}

// StyleTransferAdaIn вызывает /style_transfer_adain с двумя файлами
func (c *Client) StyleTransferAdaIn(ctx context.Context, contentData, styleData []byte, contentName, styleName string) ([]byte, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Content
	contentPart, _ := writer.CreateFormFile("content", contentName)
	contentPart.Write(contentData)

	// Style
	stylePart, _ := writer.CreateFormFile("style", styleName)
	stylePart.Write(styleData)

	writer.Close()

	req, _ := http.NewRequestWithContext(ctx, "POST", MLServiceURL+"/style_transfer_adain", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("style transfer failed: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
