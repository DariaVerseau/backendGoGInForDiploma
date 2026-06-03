package repositories

import (
	"context"
	"errors"
	"log"
	"moduleExample/web-service-gin/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Определяем кастомную ошибку
var ErrNotFound = errors.New("user not found")

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create теперь возвращает ID созданного пользователя
func (r *UserRepository) Create(ctx context.Context, email, passwordHash string) (int64, error) {
	var id int64
	err := r.db.QueryRow(ctx,
		"INSERT INTO users (email, password) VALUES ($1, $2) RETURNING id",
		email, passwordHash).Scan(&id)
	if err != nil {
		log.Printf("❌ Ошибка создания пользователя: %v", err)
		return 0, err
	}
	log.Printf("✅ Создан пользователь: ID=%d, email=%s", id, email)
	return id, nil
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	log.Printf("Запрос пользователя: %s", email)

	var user models.User
	err := r.db.QueryRow(ctx,
		"SELECT id, email, password FROM users WHERE email = $1",
		email).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		log.Printf("❌ Ошибка поиска: %v", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	log.Printf("✅ Найден пользователь: ID=%d", user.ID)
	return &user, nil
}

// Дополнительный метод для получения пользователя по ID
func (r *UserRepository) FindByID(ctx context.Context, id int64) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx,
		"SELECT id, email, password FROM users WHERE id = $1",
		id).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// internal/repositories/user_repository.go

// UpdatePassword - обновление пароля пользователя
func (r *UserRepository) UpdatePassword(ctx context.Context, userID int64, newPasswordHash string) error {
	query := `UPDATE users SET password = $1 WHERE id = $2`
	_, err := r.db.Exec(ctx, query, newPasswordHash, userID)
	return err
}

// GetUserByID - получение пользователя по ID
func (r *UserRepository) GetUserByID(ctx context.Context, userID int64) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx,
		"SELECT id, email, password FROM users WHERE id = $1",
		userID).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}
