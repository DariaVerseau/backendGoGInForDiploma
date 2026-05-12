// internal/handlers/context.go
package handlers

import (
	"context"

	"github.com/gin-gonic/gin"
)

// GetUserID получает user_id из Gin контекста
func GetUserID(c *gin.Context) int {
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

// GetContextForML возвращает контекст для ML-запросов (без токена)
func GetContextForML(c *gin.Context) context.Context {
	return c.Request.Context() // Просто возвращаем контекст без токена
}
