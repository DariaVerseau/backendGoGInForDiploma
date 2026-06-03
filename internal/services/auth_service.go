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

// Register теперь возвращает ID созданного пользователя
func (s *AuthService) Register(ctx context.Context, email, password string) (int64, error) {
	// Проверка, что email свободен
	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, repositories.ErrNotFound) {
		return 0, err
	}
	if existing != nil {
		return 0, errors.New("email уже зарегистрирован")
	}

	// Хэшируем пароль
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	// Сохраняем в БД и получаем ID
	userID, err := s.userRepo.Create(ctx, email, string(hash))
	if err != nil {
		return 0, err
	}

	return userID, nil
}

// Login теперь возвращает токен и ID пользователя
func (s *AuthService) Login(email, password string) (string, int64, error) {
	user, err := s.userRepo.FindByEmail(context.Background(), email)
	if err != nil || user == nil {
		return "", 0, ErrInvalidCredentials
	}

	// Сравниваем хэш пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		log.Printf("❌ Пароль неверен для email: %s", email)
		return "", 0, ErrInvalidCredentials
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		return "", 0, err
	}

	// 🔧 ИСПРАВЛЕНО: приводим int к int64
	return token, int64(user.ID), nil
}

// GetUserIDByEmail - вспомогательный метод для получения ID по email
func (s *AuthService) GetUserIDByEmail(email string) (int64, error) {
	user, err := s.userRepo.FindByEmail(context.Background(), email)
	if err != nil {
		return 0, err
	}
	// 🔧 ИСПРАВЛЕНО: приводим int к int64
	return int64(user.ID), nil
}

var ErrInvalidCredentials = &ServiceError{"invalid email or password"}

type ServiceError struct {
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

// ChangePassword - смена пароля пользователя
func (s *AuthService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
    // 1. Получаем пользователя из БД
    user, err := s.userRepo.GetUserByID(ctx, userID)
    if err != nil {
        return errors.New("пользователь не найден")
    }

    // 2. Проверяем старый пароль
    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
        return errors.New("неверный текущий пароль")
    }

    // 3.  ПРОВЕРКА: новый пароль не должен совпадать со старым
    if oldPassword == newPassword {
        return errors.New("новый пароль должен отличаться от текущего")
    }

    // 4. Проверяем длину нового пароля
    if len(newPassword) < 6 {
        return errors.New("новый пароль должен содержать минимум 6 символов")
    }

    // 5. Хешируем новый пароль
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
    if err != nil {
        return err
    }

    // 6. Обновляем пароль
    return s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword))
}
