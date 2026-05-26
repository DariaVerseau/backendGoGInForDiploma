package handlers

import (
	"context"
	"log"
	"moduleExample/web-service-gin/internal/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ImageHandler struct {
	imageService *services.ImageService
}

func NewImageHandler(svc *services.ImageService) *ImageHandler {
	return &ImageHandler{imageService: svc}
}

func (h *ImageHandler) GetImage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(401, gin.H{"error": "пользователь не авторизован"})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		c.AbortWithStatusJSON(500, gin.H{"error": "неверный тип ID пользователя"})
		return
	}

	imageID := c.Param("id")
	if imageID == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "image ID is required"})
		return
	}

	img, err := h.imageService.GetImageByID(c.Request.Context(), imageID, int(userIDInt64))
	if err != nil {
		if err.Error() == "изображение не найдено" || err.Error() == "доступ запрещён" {
			c.AbortWithStatusJSON(404, gin.H{"error": "изображение не найдено"})
			return
		}
		c.AbortWithStatusJSON(500, gin.H{"error": "внутренняя ошибка"})
		return
	}

	log.Printf("Получено изображение из БД: %+v", img)
	c.JSON(200, img)
}

func (h *ImageHandler) GetImages(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(401, gin.H{"error": "пользователь не авторизован"})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		c.AbortWithStatusJSON(500, gin.H{"error": "неверный тип ID пользователя"})
		return
	}

	images, err := h.imageService.GetImagesByUser(c.Request.Context(), int(userIDInt64))
	if err != nil {
		log.Printf("Ошибка получения изображений: %v", err)
		c.AbortWithStatusJSON(500, gin.H{"error": "ошибка сервера"})
		return
	}

	c.JSON(200, images)
}

func (h *ImageHandler) PostImage(c *gin.Context) {
	c.AbortWithStatusJSON(400, gin.H{"error": "используйте POST /images с multipart/form-data для загрузки файлов"})
}

func (h *ImageHandler) UploadImage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(401, gin.H{"error": "пользователь не авторизован"})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		c.AbortWithStatusJSON(500, gin.H{"error": "неверный тип ID пользователя"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "файл 'image' обязателен"})
		return
	}

	title := c.PostForm("title")
	style := c.PostForm("style")

	img, err := h.imageService.UploadImage(c.Request.Context(), int(userIDInt64), file, title, style)
	if err != nil {
		log.Printf("Ошибка загрузки: %v", err)
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(201, img)
}

// ========== НОВЫЕ МЕТОДЫ ДЛЯ ML ОПЕРАЦИЙ ==========

// Upscale - улучшение качества / увеличение разрешения
func (h *ImageHandler) Upscale(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(401, gin.H{"error": "пользователь не авторизован"})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		c.AbortWithStatusJSON(500, gin.H{"error": "неверный тип ID пользователя"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "файл 'image' обязателен"})
		return
	}

	scaleStr := c.PostForm("scale")
	scale := 2
	if scaleStr == "4" {
		scale = 4
	}

	// ✅ Передаём токен в контекст
	authHeader := c.GetHeader("Authorization")
	ctx := context.WithValue(c.Request.Context(), "token", authHeader)

	result, err := h.imageService.Upscale(ctx, int(userIDInt64), file, scale)
	if err != nil {
		log.Printf("Ошибка upscale: %v", err)
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, result)
}

// Enhance - улучшение изображения
func (h *ImageHandler) Enhance(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(401, gin.H{"error": "пользователь не авторизован"})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		c.AbortWithStatusJSON(500, gin.H{"error": "неверный тип ID пользователя"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "файл 'image' обязателен"})
		return
	}

	// ✅ Добавляем параметры со значениями по умолчанию
	fidelityWeight := 0.5
	postprocess := true

	ctx := c.Request.Context()
	result, err := h.imageService.Enhance(ctx, int(userIDInt64), file, fidelityWeight, postprocess)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, result)
}

// PostProcess - постобработка изображения
func (h *ImageHandler) PostProcess(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(401, gin.H{"error": "пользователь не авторизован"})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		c.AbortWithStatusJSON(500, gin.H{"error": "неверный тип ID пользователя"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "файл 'image' обязателен"})
		return
	}

	// Получаем параметры постобработки
	sharpness := c.PostForm("sharpness")
	contrast := c.PostForm("contrast")
	brightness := c.PostForm("brightness")
	denoise := c.PostForm("denoise")

	authHeader := c.GetHeader("Authorization")
	ctx := context.WithValue(c.Request.Context(), "token", authHeader)

	// Собираем параметры
	params := make(map[string]string)
	if sharpness != "" {
		params["sharpness"] = sharpness
	}
	if contrast != "" {
		params["contrast"] = contrast
	}
	if brightness != "" {
		params["brightness"] = brightness
	}
	if denoise != "" {
		params["denoise"] = denoise
	}

	result, err := h.imageService.PostProcessWithParams(ctx, int(userIDInt64), file, params)
	if err != nil {
		log.Printf("Ошибка postprocess: %v", err)
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, result)
}

// StyleTransfer - перенос стиля
func (h *ImageHandler) StyleTransfer(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(401, gin.H{"error": "пользователь не авторизован"})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		c.AbortWithStatusJSON(500, gin.H{"error": "неверный тип ID пользователя"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "файл 'image' обязателен"})
		return
	}

	style := c.PostForm("style")
	if style == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "стиль обязателен"})
		return
	}

	// Читаем ДОПОЛНИТЕЛЬНЫЕ параметры
	alphaStr := c.DefaultPostForm("alpha", "1.0")
	preserveColorStr := c.DefaultPostForm("preserve_color", "false")

	alpha, _ := strconv.ParseFloat(alphaStr, 32)
	preserveColor := preserveColorStr == "true"

	authHeader := c.GetHeader("Authorization")
	ctx := context.WithValue(c.Request.Context(), "token", authHeader)

	// Передаём ВСЕ параметры
	result, err := h.imageService.StyleTransfer(
		ctx,
		int(userIDInt64),
		file,
		style,
		float32(alpha),
		preserveColor,
	)
	if err != nil {
		log.Printf("Ошибка style_transfer: %v", err)
		c.AbortWithStatusJSON(500, gin.H{"error": err.Error()}) // ← 500 вместо 400
		return
	}

	c.JSON(200, result)
}

// BasicStyleTransfer - базовый перенос стиля
func (h *ImageHandler) BasicStyleTransfer(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(401, gin.H{"error": "пользователь не авторизован"})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		c.AbortWithStatusJSON(500, gin.H{"error": "неверный тип ID пользователя"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "файл 'image' обязателен"})
		return
	}

	style := c.PostForm("style")
	if style == "" {
		c.AbortWithStatusJSON(400, gin.H{"error": "стиль обязателен"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	ctx := context.WithValue(c.Request.Context(), "token", authHeader)

	result, err := h.imageService.BasicStyleTransfer(ctx, int(userIDInt64), file, style)
	if err != nil {
		log.Printf("Ошибка basic_style_transfer: %v", err)
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, result)
}

// Colorize - раскрашивание изображения
func (h *ImageHandler) Colorize(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(401, gin.H{"error": "пользователь не авторизован"})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		c.AbortWithStatusJSON(500, gin.H{"error": "неверный тип ID пользователя"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": "файл 'image' обязателен"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	ctx := context.WithValue(c.Request.Context(), "token", authHeader)

	result, err := h.imageService.Colorize(ctx, int(userIDInt64), file)
	if err != nil {
		log.Printf("Ошибка colorize: %v", err)
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, result)
}
