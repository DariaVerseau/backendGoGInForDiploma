package handlers

import (
	"io"
	"moduleExample/web-service-gin/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MLHandler struct {
	mlService *services.MLService
}

func NewMLHandler(mlService *services.MLService) *MLHandler {
	return &MLHandler{mlService: mlService}
}

func (h *MLHandler) Upscale(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing image"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer src.Close()

	data := make([]byte, file.Size)
	_, err = src.Read(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	resultPath, err := h.mlService.UpscaleImage(c.Request.Context(), data, file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result_url": "/uploads/" + resultPath})
}

func (h *MLHandler) EnhanceFace(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing image"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer src.Close()

	data := make([]byte, file.Size)
	_, err = src.Read(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	resultPath, err := h.mlService.EnhanceFace(c.Request.Context(), data, file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result_url": "/uploads/" + resultPath})
}

func (h *MLHandler) Colorize(c *gin.Context) {
	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing image"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open file"})
		return
	}
	defer src.Close()

	// Читаем данные в переменную `data`
	data, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	resultURL, err := h.mlService.ColorizeImage(c.Request.Context(), data, file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result_url": resultURL})
}

func (h *MLHandler) StyleTransferAdaIN(c *gin.Context) {
	contentFile, err := c.FormFile("content")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing content image"})
		return
	}

	styleFile, err := c.FormFile("style")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing style image"})
		return
	}

	// Читаем content
	contentSrc, err := contentFile.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open content"})
		return
	}
	defer contentSrc.Close()
	contentData, err := io.ReadAll(contentSrc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read content"})
		return
	}

	// Читаем style
	styleSrc, err := styleFile.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to open style"})
		return
	}
	defer styleSrc.Close()
	styleData, err := io.ReadAll(styleSrc)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read style"})
		return
	}

	resultURL, err := h.mlService.ApplyStyle(
		c.Request.Context(),
		contentData,
		styleData,
		contentFile.Filename,
		styleFile.Filename,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"result_url": resultURL})
}
