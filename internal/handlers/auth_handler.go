package handlers

import (
	"log"
	"moduleExample/web-service-gin/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=6"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Register validation error: %v", err)
		c.AbortWithStatusJSON(400, gin.H{"error": "некорректные данные"})
		return
	}

	// Регистрируем пользователя и получаем его ID
	userID, err := h.authService.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	// Генерируем токен для нового пользователя
	token, _, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		c.AbortWithStatusJSON(500, gin.H{"error": "failed to generate token"})
		return
	}

	// Возвращаем и токен, и user_id
	c.JSON(201, gin.H{
		"token":   token,
		"user_id": userID,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid input"})
		return
	}

	token, userID, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем и токен, и user_id
	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"user_id": userID,
	})
}