package handlers

import (
	"log"
	"moduleExample/web-service-gin/internal/models"
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

func (h *MLHandler) Upscale(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "пользователь не авторизован"})
		return
	}

	scaleStr := c.DefaultPostForm("scale", "4")
	scale, err := strconv.Atoi(scaleStr)
	if err != nil || (scale != 2 && scale != 4) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "scale must be 2 or 4"})
		return
	}

	imageID := c.PostForm("image_id")

	if imageID != "" {
		log.Printf("Upscale by ID: %s, scale: %d", imageID, scale)
		result, err := h.imageService.UpscaleByID(c.Request.Context(), userID, imageID, scale)
		if err != nil {
			log.Printf("UpscaleByID error: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	// Fallback: обработка с загруженным файлом
	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "поле 'image' или 'image_id' обязательно"})
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

	fidelityWeightStr := c.DefaultPostForm("fidelity_weight", "0.5")
	postprocessStr := c.DefaultPostForm("postprocess", "true")

	fidelityWeight, _ := strconv.ParseFloat(fidelityWeightStr, 64)
	postprocess := postprocessStr == "true"

	imageID := c.PostForm("image_id")

	if imageID != "" {
		log.Printf("Enhance by ID: %s", imageID)
		result, err := h.imageService.EnhanceByID(c.Request.Context(), userID, imageID, fidelityWeight, postprocess)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "поле 'image' или 'image_id' обязательно"})
		return
	}

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

	// Собираем параметры
	params := make(map[string]string)
	if sharpness := c.PostForm("sharpness"); sharpness != "" {
		params["sharpness"] = sharpness
	}
	if contrast := c.PostForm("contrast"); contrast != "" {
		params["contrast"] = contrast
	}
	if brightness := c.PostForm("brightness"); brightness != "" {
		params["brightness"] = brightness
	}
	if denoise := c.PostForm("denoise"); denoise != "" {
		params["denoise"] = denoise
	}

	imageID := c.PostForm("image_id")

	if imageID != "" {
		log.Printf("PostProcess by ID: %s", imageID)
		result, err := h.imageService.PostProcessByID(c.Request.Context(), userID, imageID, params)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "поле 'image' или 'image_id' обязательно"})
		return
	}

	var img *models.Image
	if len(params) > 0 {
		img, err = h.imageService.PostProcessWithParams(c.Request.Context(), userID, file, params)
	} else {
		img, err = h.imageService.PostProcess(c.Request.Context(), userID, file)
	}
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

	styleName := c.DefaultPostForm("style", "vangogh")
	alphaStr := c.DefaultPostForm("alpha", "1.0")
	preserveColorStr := c.DefaultPostForm("preserve_color", "false")

	var alpha float64 = 1.0
	if parsedAlpha, err := strconv.ParseFloat(alphaStr, 32); err == nil {
		alpha = parsedAlpha
	}
	preserveColor := preserveColorStr == "true"

	imageID := c.PostForm("image_id")

	if imageID != "" {
		log.Printf("StyleTransfer by ID: %s, style: %s", imageID, styleName)
		result, err := h.imageService.StyleTransferByID(c.Request.Context(), userID, imageID, styleName, float32(alpha), preserveColor)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "поле 'image' или 'image_id' обязательно"})
		return
	}

	img, err := h.imageService.StyleTransfer(c.Request.Context(), userID, file, styleName, float32(alpha), preserveColor)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, img)
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
