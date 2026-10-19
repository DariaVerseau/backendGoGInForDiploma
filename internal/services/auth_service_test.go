package services

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DariaVerseau/backendGoGInForDiploma/internal/apperrors"
	"github.com/DariaVerseau/backendGoGInForDiploma/internal/auth"
	"github.com/DariaVerseau/backendGoGInForDiploma/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	findByEmailFunc    func(ctx context.Context, email string) (*models.User, error)
	createFunc         func(ctx context.Context, email, passwordHash string) (int64, error)
	getUserByIDFunc    func(ctx context.Context, id int64) (*models.User, error)
	updatePasswordFunc func(ctx context.Context, userID int64, passwordHash string) error
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	return m.findByEmailFunc(ctx, email)
}

func (m *mockUserRepo) Create(ctx context.Context, email, passwordHash string) (int64, error) {
	return m.createFunc(ctx, email, passwordHash)
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	return m.getUserByIDFunc(ctx, id)
}

func (m *mockUserRepo) UpdatePassword(ctx context.Context, userID int64, passwordHash string) error {
	return m.updatePasswordFunc(ctx, userID, passwordHash)
}

func TestAuthService_Register_Success(t *testing.T) {
	tm, _ := auth.NewTokenManager("test_secret_at_least_32_characters_long")
	repo := &mockUserRepo{
		findByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return nil, apperrors.ErrNotFound // ← вот это
		},
		createFunc: func(ctx context.Context, email, passwordHash string) (int64, error) {
			return 1, nil
		},
	}
	svc := NewAuthService(repo, tm)

	token, userID, err := svc.Register(context.Background(), "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if token == "" {
		t.Error("token is empty")
	}
	if userID != 1 {
		t.Errorf("userID = %d, want 1", userID)
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	tm, _ := auth.NewTokenManager("test_secret_at_least_32_characters_long")
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	repo := &mockUserRepo{
		findByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return &models.User{ID: 1, Email: email, PasswordHash: string(hash)}, nil
		},
	}
	svc := NewAuthService(repo, tm)

	_, _, err := svc.Login(context.Background(), "test@example.com", "wrong")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthService_GetUserIDByEmail_Success(t *testing.T) {
	tm, _ := auth.NewTokenManager("test_secret_at_least_32_characters_long")
	repo := &mockUserRepo{
		findByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return &models.User{ID: 42, Email: email}, nil
		},
	}
	svc := NewAuthService(repo, tm)

	userID, err := svc.GetUserIDByEmail("test@example.com")
	if err != nil {
		t.Fatalf("GetUserIDByEmail() error = %v", err)
	}
	if userID != 42 {
		t.Errorf("userID = %d, want 42", userID)
	}
}

func TestAuthService_ChangePassword_ShortNewPassword(t *testing.T) {
	tm, _ := auth.NewTokenManager("test_secret_at_least_32_characters_long")
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	repo := &mockUserRepo{
		getUserByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
			return &models.User{ID: id, PasswordHash: string(oldHash)}, nil
		},
	}
	svc := NewAuthService(repo, tm)

	err := svc.ChangePassword(context.Background(), 1, "oldpass", "abc")
	if err == nil {
		t.Error("ChangePassword() expected error for short password, got nil")
	}
}

func TestAuthService_ChangePassword_SamePassword(t *testing.T) {
	tm, _ := auth.NewTokenManager("test_secret_at_least_32_characters_long")
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("same"), bcrypt.DefaultCost)
	repo := &mockUserRepo{
		getUserByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
			return &models.User{ID: id, PasswordHash: string(oldHash)}, nil
		},
	}
	svc := NewAuthService(repo, tm)

	err := svc.ChangePassword(context.Background(), 1, "same", "same")
	if err == nil {
		t.Error("ChangePassword() expected error for same password, got nil")
	}
}

func TestAuthService_ChangePassword_WrongOldPassword(t *testing.T) {
	tm, _ := auth.NewTokenManager("test_secret_at_least_32_characters_long")
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	repo := &mockUserRepo{
		getUserByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
			return &models.User{ID: id, PasswordHash: string(oldHash)}, nil
		},
	}
	svc := NewAuthService(repo, tm)

	err := svc.ChangePassword(context.Background(), 1, "wrong", "newpass")
	if err == nil {
		t.Error("ChangePassword() expected error for wrong old password, got nil")
	}
}

func TestAuthService_ChangePassword_Success(t *testing.T) {
	tm, _ := auth.NewTokenManager("test_secret_at_least_32_characters_long")
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	repo := &mockUserRepo{
		getUserByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
			return &models.User{ID: id, PasswordHash: string(oldHash)}, nil
		},
		updatePasswordFunc: func(ctx context.Context, userID int64, passwordHash string) error {
			return nil
		},
	}
	svc := NewAuthService(repo, tm)

	err := svc.ChangePassword(context.Background(), 1, "oldpass", "newpass")
	if err != nil {
		t.Fatalf("ChangePassword() error = %v", err)
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	tm, _ := auth.NewTokenManager("test_secret_at_least_32_characters_long")
	repo := &mockUserRepo{
		findByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return nil, apperrors.ErrNotFound
		},
	}
	svc := NewAuthService(repo, tm)

	_, _, err := svc.Login(context.Background(), "test@example.com", "password123")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	tm, _ := auth.NewTokenManager("test_secret_at_least_32_characters_long")
	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	repo := &mockUserRepo{
		findByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return &models.User{ID: 1, Email: email, PasswordHash: string(hash)}, nil
		},
	}
	svc := NewAuthService(repo, tm)

	token, userID, err := svc.Login(context.Background(), "test@example.com", "correct")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if token == "" {
		t.Error("token is empty")
	}
	if userID != 1 {
		t.Errorf("userID = %d, want 1", userID)
	}
}

func TestAuthService_Register_EmailTaken(t *testing.T) {
	tm, _ := auth.NewTokenManager("test_secret_at_least_32_characters_long")
	repo := &mockUserRepo{
		findByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return &models.User{ID: 1, Email: email}, nil // email занят
		},
	}
	svc := NewAuthService(repo, tm)

	_, _, err := svc.Register(context.Background(), "test@example.com", "password123")
	if err == nil {
		t.Error("Register() expected error for taken email, got nil")
	}
}

func TestServiceError_Error(t *testing.T) {
	e := &ServiceError{Message: "test error"}
	if e.Error() != "test error" {
		t.Errorf("Error() = %q, want %q", e.Error(), "test error")
	}
}

func TestAuthService_Register_RepoError(t *testing.T) {
	tm, _ := auth.NewTokenManager("test_secret_at_least_32_characters_long")
	repo := &mockUserRepo{
		findByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return nil, errors.New("db connection lost") // ← не ErrNotFound
		},
	}
	svc := NewAuthService(repo, tm)

	_, _, err := svc.Register(context.Background(), "test@example.com", "password123")
	if err == nil {
		t.Error("Register() expected error for repo failure, got nil")
	}
}

func TestAuthService_GetUserIDByEmail_NotFound(t *testing.T) {
	tm, _ := auth.NewTokenManager("test_secret_at_least_32_characters_long")
	repo := &mockUserRepo{
		findByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return nil, apperrors.ErrNotFound
		},
	}
	svc := NewAuthService(repo, tm)

	_, err := svc.GetUserIDByEmail("test@example.com")
	if err == nil {
		t.Error("GetUserIDByEmail() expected error, got nil")
	}
}

func TestAuthService_Register_CreateError(t *testing.T) {
	tm, _ := auth.NewTokenManager("test_secret_at_least_32_characters_long")
	repo := &mockUserRepo{
		findByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
			return nil, apperrors.ErrNotFound
		},
		createFunc: func(ctx context.Context, email, passwordHash string) (int64, error) {
			return 0, errors.New("db connection lost")
		},
	}
	svc := NewAuthService(repo, tm)

	_, _, err := svc.Register(context.Background(), "test@example.com", "password123")
	if err == nil {
		t.Error("Register() expected error for create failure, got nil")
	}
}

func TestAuthService_ChangePassword_UpdateError(t *testing.T) {
	tm, _ := auth.NewTokenManager("test_secret_at_least_32_characters_long")
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	repo := &mockUserRepo{
		getUserByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
			return &models.User{ID: id, PasswordHash: string(oldHash)}, nil
		},
		updatePasswordFunc: func(ctx context.Context, userID int64, passwordHash string) error {
			t.Log("UpdatePassword CALLED")
			return errors.New("db connection lost")
		},
	}
	svc := NewAuthService(repo, tm)

	err := svc.ChangePassword(context.Background(), 1, "oldpass", "newpass")
	if err == nil {
		t.Error("ChangePassword() expected error for update failure, got nil")
	}
}

func TestAuthService_ChangePassword_BcryptError(t *testing.T) {
	tm, _ := auth.NewTokenManager("test_secret_at_least_32_characters_long")
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("oldpass"), bcrypt.DefaultCost)
	repo := &mockUserRepo{
		getUserByIDFunc: func(ctx context.Context, id int64) (*models.User, error) {
			return &models.User{ID: id, PasswordHash: string(oldHash)}, nil
		},
	}
	svc := NewAuthService(repo, tm)

	// bcrypt не поддерживает пароль длиннее 72 байт
	longPassword := strings.Repeat("a", 100)
	err := svc.ChangePassword(context.Background(), 1, "oldpass", longPassword)
	if err == nil {
		t.Error("ChangePassword() expected error for long password, got nil")
	}
}
