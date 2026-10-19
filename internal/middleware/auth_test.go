package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DariaVerseau/backendGoGInForDiploma/internal/auth"
	"github.com/gin-gonic/gin"
)

func TestAuthMiddleware_InvalidCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tm, _ := auth.NewTokenManager("test_secret_at_least_32_characters_long")

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{"no header", "", http.StatusUnauthorized},
		{"invalid format", "Token abc", http.StatusUnauthorized},
		{"invalid token", "Bearer invalid", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(AuthMiddleware(tm))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tm, _ := auth.NewTokenManager("test_secret_at_least_32_characters_long")
	token, _ := tm.GenerateToken(int64(42))

	router := gin.New()
	router.Use(AuthMiddleware(tm))
	router.GET("/test", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Body.String() != `{"user_id":42}` {
		t.Errorf("body = %s, want user_id=42", w.Body.String())
	}
}
