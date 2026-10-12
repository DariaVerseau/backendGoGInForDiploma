package handlers

import (
	"log"
	"net/http"

	"github.com/DariaVerseau/backendGoGInForDiploma/internal/services"

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
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "некорректные данные"})
		return
	}

	// Регистрируем пользователя и получаем его ID
	token, userID, err := h.authService.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
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

	token, userID, err := h.authService.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"user_id": userID,
	})
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userIDInterface, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(401, gin.H{"error": "пользователь не авторизован"})
		return
	}

	userID, ok := userIDInterface.(int64)
	if !ok {
		c.AbortWithStatusJSON(500, gin.H{"error": "неверный тип ID пользователя"})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.ChangePassword(
		c.Request.Context(),
		userID,
		req.OldPassword,
		req.NewPassword,
	)
	if err != nil {
		c.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "пароль успешно изменён"})
}
