package handlers

import (
	"moduleExample/web-service-gin/internal/services"

	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	imageService *services.ImageService // ← зависимость от сервиса
}

func NewImageHandler(svc *services.ImageService) *ImageHandler {
	return &ImageHandler{imageService: svc}
}

// internal/handlers/image_handler.go
func (h *ImageHandler) GetImages(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(401, gin.H{"error": "пользователь не авторизован"})
		return
	}

	images, err := h.imageService.GetImagesByUser(c.Request.Context(), userID.(int))
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "ошибка сервера"})
		return
	}

	c.JSON(200, images)
}

func (h *ImageHandler) PostImage(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var req struct {
		Title string `json:"title" binding:"required"`
		Style string `json:"style" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "title and style are required"})
		return
	}

	img, err := h.imageService.CreateImage(c.Request.Context(), userID.(int), req.Title, req.Style)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}
	c.JSON(201, img)
}

func (h *ImageHandler) UploadImage(c *gin.Context) {
	userID, _ := c.Get("user_id")

	// Получаем файл
	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "файл 'image' обязателен"})
		return
	}

	// Получаем метаданные из формы
	title := c.PostForm("title")
	style := c.PostForm("style")

	// Передаём в сервис
	img, err := h.imageService.UploadImage(c.Request.Context(), userID.(int), file, title, style)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, img)
}
