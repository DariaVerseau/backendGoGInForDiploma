package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/DariaVerseau/backendGoGInForDiploma/internal/services"
	"github.com/google/uuid"

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
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "пользователь не авторизован"})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "неверный тип ID пользователя"})
		return
	}

	imageIDStr := c.Param("id")
	if imageIDStr == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "image ID is required"})
		return
	}

	imageID, err := uuid.Parse(imageIDStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid image ID"})
		return
	}

	img, err := h.imageService.GetImageByID(c.Request.Context(), imageID, userIDInt64)
	if err != nil {
		if err.Error() == "изображение не найдено" || err.Error() == "доступ запрещён" {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "изображение не найдено"})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "внутренняя ошибка"})
		return
	}

	log.Printf("Получено изображение из БД: %+v", img)
	c.JSON(http.StatusOK, img)
}

func (h *ImageHandler) GetImages(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "пользователь не авторизован"})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "неверный тип ID пользователя"})
		return
	}

	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	images, err := h.imageService.GetImagesByUserPaginated(
		c.Request.Context(),
		userIDInt64,
		limit,
		offset,
	)
	if err != nil {
		log.Printf("Ошибка получения изображений: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "ошибка сервера"})
		return
	}

	total, err := h.imageService.GetImagesCount(c.Request.Context(), userIDInt64)
	if err != nil {
		log.Printf("Ошибка подсчёта изображений: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "ошибка сервера"})
		return
	}

	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"items":       images,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

func (h *ImageHandler) PostImage(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "используйте POST /images с multipart/form-data для загрузки файлов"})
}

func (h *ImageHandler) UploadImage(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "пользователь не авторизован"})
		return
	}

	userIDInt64, ok := userID.(int64)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "неверный тип ID пользователя"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "файл 'image' обязателен"})
		return
	}

	title := c.PostForm("title")
	style := c.PostForm("style")

	img, err := h.imageService.UploadImage(c.Request.Context(), userIDInt64, file, title, style)
	if err != nil {
		log.Printf("Ошибка загрузки: %v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, img)
}
