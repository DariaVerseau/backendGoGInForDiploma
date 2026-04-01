package services

import (
	"context"
	"errors"
	"log"
	"moduleExample/web-service-gin/internal/auth"
	"moduleExample/web-service-gin/internal/repositories"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *repositories.UserRepository
}

func NewAuthService(userRepo *repositories.UserRepository) *AuthService {
	return &AuthService{userRepo: userRepo}
}

func (s *AuthService) Register(ctx context.Context, email, password string) error {
	// Проверка, что email свободен
	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, repositories.ErrNotFound) {
		return err // ошибка БД — прокидываем наверх
	}
	if existing != nil {
		return errors.New("email уже зарегистрирован")
	}

	// Хэшируем пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Сохраняем в БД
	return s.userRepo.Create(context.Background(), email, string(hash))
}

func (s *AuthService) Login(email, password string) (string, error) {
	user, err := s.userRepo.FindByEmail(context.Background(), email)
	if err != nil || user == nil {
		return "", ErrInvalidCredentials
	}

	// Сравниваем хэш пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		log.Printf("❌ Пароль неверен для email: %s", email)
		return "", ErrInvalidCredentials
	}

	return auth.GenerateToken(user.ID)
}

var ErrInvalidCredentials = &ServiceError{"invalid email or password"}

type ServiceError struct {
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}
