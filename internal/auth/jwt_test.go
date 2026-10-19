package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenManager_GenerateAndValidate(t *testing.T) {
	secret := "test_secret_at_least_32_characters_long"
	tm, err := NewTokenManager(secret)
	if err != nil {
		t.Fatalf("NewTokenManager() error = %v", err)
	}

	userID := int64(42)
	token, err := tm.GenerateToken(userID)
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	claims, err := tm.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken() error = %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("UserID = %d, want %d", claims.UserID, userID)
	}
}

func TestTokenManager_ExpiredToken(t *testing.T) {
	secret := "test_secret_at_least_32_characters_long"
	tm, _ := NewTokenManager(secret)

	claims := &Claims{
		UserID: 1,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))

	_, err := tm.ValidateToken(tokenString)
	if err == nil {
		t.Error("ValidateToken() expected error for expired token, got nil")
	}
}

func TestTokenManager_WrongAlgorithm(t *testing.T) {
	secret := "test_secret_at_least_32_characters_long"
	tm, _ := NewTokenManager(secret)

	// Токен с alg=none (alg confusion)
	token := jwt.NewWithClaims(jwt.SigningMethodNone, &Claims{UserID: 1})
	tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

	_, err := tm.ValidateToken(tokenString)
	if err == nil {
		t.Error("ValidateToken() expected error for alg=none, got nil")
	}
}

func TestNewTokenManager_ShortSecret(t *testing.T) {
	_, err := NewTokenManager("short")
	if err == nil {
		t.Error("NewTokenManager() expected error for short secret, got nil")
	}
}
