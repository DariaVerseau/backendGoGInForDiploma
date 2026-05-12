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

// getUserIDFromContext получает user_id из Gin контекста
func getUserIDFromContext(c *gin.Context) int {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	switch v := userID.(type) {
	case int:
		return v
	case int64:
		return int(v)
	default:
		return 0
	}
}

func (h *MLHandler) Process(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "пользователь не авторизован"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "поле 'image' обязательно"})
		return
	}

	styleName := c.DefaultPostForm("style", "vangogh")

	img, err := h.imageService.Process(c.Request.Context(), userID, file, styleName)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, img)
}

func (h *MLHandler) Upscale(c *gin.Context) {
	userID := getUserIDFromContext(c)
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

func (h *MLHandler) Enhance(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "пользователь не авторизован"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "поле 'image' обязательно"})
		return
	}

	fidelityWeightStr := c.DefaultPostForm("fidelity_weight", "0.5")
	postprocessStr := c.DefaultPostForm("postprocess", "true")

	fidelityWeight, _ := strconv.ParseFloat(fidelityWeightStr, 64)
	postprocess := postprocessStr == "true"

	// ✅ Просто используем c.Request.Context() - без токена
	img, err := h.imageService.Enhance(c.Request.Context(), userID, file, fidelityWeight, postprocess)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, img)
}

func (h *MLHandler) PostProcess(c *gin.Context) {
	userID := getUserIDFromContext(c)
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

func (h *MLHandler) StyleTransfer(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "пользователь не авторизован"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "поле 'image' обязательно"})
		return
	}

	styleName := c.DefaultPostForm("style", "vangogh")

	img, err := h.imageService.StyleTransfer(c.Request.Context(), userID, file, styleName)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, img)
}

func (h *MLHandler) Colorize(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "пользователь не авторизован"})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "поле 'image' обязательно"})
		return
	}

	img, err := h.imageService.Colorize(c.Request.Context(), userID, file)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, img)
}
