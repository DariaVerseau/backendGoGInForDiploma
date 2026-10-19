package services

import (
	"context"
	"errors"
	"log"

	"github.com/DariaVerseau/backendGoGInForDiploma/internal/apperrors"
	"github.com/DariaVerseau/backendGoGInForDiploma/internal/auth"
	"github.com/DariaVerseau/backendGoGInForDiploma/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	Create(ctx context.Context, email, passwordHash string) (int64, error)
	GetUserByID(ctx context.Context, id int64) (*models.User, error)
	UpdatePassword(ctx context.Context, userID int64, passwordHash string) error
}

type AuthService struct {
	userRepo     UserRepository
	tokenManager *auth.TokenManager
}

func NewAuthService(userRepo UserRepository, tokenManager *auth.TokenManager) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		tokenManager: tokenManager,
	}
}

func (s *AuthService) Register(ctx context.Context, email, password string) (string, int64, error) {
	existing, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return "", 0, err
	}
	if existing != nil {
		return "", 0, errors.New("email already regist")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", 0, err
	}

	userID, err := s.userRepo.Create(ctx, email, string(hash))
	if err != nil {
		return "", 0, err
	}

	token, err := s.tokenManager.GenerateToken(userID)
	if err != nil {
		return "", 0, err
	}

	return token, userID, nil
}

// Login возвращает токен и ID пользователя
func (s *AuthService) Login(ctx context.Context, email, password string) (string, int64, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil || user == nil {
		return "", 0, ErrInvalidCredentials
	}

	// Сравниваем хэш пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		log.Printf("Пароль неверен для email: %s", email)
		return "", 0, ErrInvalidCredentials
	}

	token, err := s.tokenManager.GenerateToken(int64(user.ID))
	if err != nil {
		return "", 0, err
	}

	return token, int64(user.ID), nil
}

// GetUserIDByEmail - вспомогательный метод для получения ID по email
func (s *AuthService) GetUserIDByEmail(email string) (int64, error) {
	user, err := s.userRepo.FindByEmail(context.Background(), email)
	if err != nil {
		return 0, err
	}

	return int64(user.ID), nil
}

var ErrInvalidCredentials = &ServiceError{"invalid email or password"}

type ServiceError struct {
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

func (s *AuthService) ChangePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return errors.New("пользователь не найден")
	}

	// Проверяем старый пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("неверный текущий пароль")
	}

	// новый пароль не должен совпадать со старым
	if oldPassword == newPassword {
		return errors.New("новый пароль должен отличаться от текущего")
	}

	// Проверяем длину нового пароля
	if len(newPassword) < 6 {
		return errors.New("новый пароль должен содержать минимум 6 символов")
	}

	// Хешируем новый пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Обновляем пароль
	return s.userRepo.UpdatePassword(ctx, userID, string(hashedPassword))
}
