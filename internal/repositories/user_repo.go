package repositories

import (
	"context"
	"errors"
	"log"
	"moduleExample/web-service-gin/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, email, passwordHash string) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO users (email, password) VALUES ($1, $2)",
		email, passwordHash)
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	log.Printf("🔍 Запрос пользователя: %s", email) // ← добавь эту строку

	var user models.User
	err := r.db.QueryRow(ctx,
		"SELECT id, email, password FROM users WHERE email = $1",
		email).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		log.Printf("❌ Ошибка поиска: %v", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	log.Printf("✅ Найден пользователь: ID=%d", user.ID)
	return &user, nil
}
