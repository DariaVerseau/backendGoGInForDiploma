package auth

import (
	"errors"
	"fmt"

	"time"

	"github.com/golang-jwt/jwt/v5"
)

const(
	minSecretLength = 32
	tokenTTL = 24 * time.Hour
)

type Claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret []byte
}

func NewTokenManager(secret string)(*TokenManager, error){
	if len(secret) < minSecretLength{
		return nil, fmt.Errorf("JWT secret must be at least %d characters", minSecretLength)
	}
	return &TokenManager{secret: []byte(secret)}, nil
}

func (m *TokenManager) GenerateToken(userID int64) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *TokenManager) ValidateToken(tokenString string) (*Claims, error){
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(t *jwt.Token) (interface{}, error){
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok{
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return m.secret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
	)
	if err != nil{
		return nil, fmt.Errorf("parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid{
		return nil, errors.New("invalid token")
	}
	return claims, nil
}