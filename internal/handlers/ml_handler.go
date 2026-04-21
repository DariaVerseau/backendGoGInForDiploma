package handlers

import (
	"moduleExample/web-service-gin/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MLHandler struct {
	imageService *services.ImageService
}

func NewMLHandler(imageService *services.ImageService) *MLHandler {
	return &MLHandler{imageService: imageService}
}

func (h *MLHandler) Process(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "пользователь не авторизован"})
		return
	}

	contentFile, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "поле 'image' (content) обязательно"})
		return
	}

	styleFile, err := c.FormFile("style")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "поле 'style' обязательно для NST"})
		return
	}

	img, err := h.imageService.Process(c.Request.Context(), userID, contentFile, styleFile)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, img)
}

func (h *MLHandler) Upscale(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "пользователь не авторизован"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "поле 'image' обязательно"})
		return
	}

	scaleStr := c.DefaultPostForm("scale", "4")
	scale, err := strconv.Atoi(scaleStr)
	if err != nil || (scale != 2 && scale != 4) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "scale must be 2 or 4"})
		return
	}

	img, err := h.imageService.Upscale(c.Request.Context(), userID, file, scale)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, img)
}

// Enhance — улучшение лица / качества
func (h *MLHandler) Enhance(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "пользователь не авторизован"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "поле 'image' обязательно"})
		return
	}

	img, err := h.imageService.Enhance(c.Request.Context(), userID, file)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, img)
}

// PostProcess — постобработка (шум, цвет и т.д.)
func (h *MLHandler) PostProcess(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "пользователь не авторизован"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "поле 'image' обязательно"})
		return
	}

	img, err := h.imageService.PostProcess(c.Request.Context(), userID, file)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, img)
}

// StyleTransfer — перенос стиля (требует два файла)
func (h *MLHandler) StyleTransfer(c *gin.Context) {
	userID := getUserID(c)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "пользователь не авторизован"})
		return
	}

	contentFile, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "поле 'image' (content) обязательно"})
		return
	}

	styleFile, err := c.FormFile("style")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "поле 'style' обязательно для переноса стиля"})
		return
	}

	img, err := h.imageService.StyleTransfer(c.Request.Context(), userID, contentFile, styleFile)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, img)
}

// Вспомогательная функция (должна быть в middleware/auth.go или здесь)
func getUserID(c *gin.Context) int {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	if id, ok := userID.(int); ok {
		return id
	}
	return 0
}
