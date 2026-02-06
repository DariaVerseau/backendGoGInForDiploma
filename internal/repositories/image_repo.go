package repositories

import (
	"context"
	"errors"
	"fmt"
	"moduleExample/web-service-gin/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ImageRepository struct {
	db *pgxpool.Pool
}

func NewImageRepository(db *pgxpool.Pool) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) GetAll(ctx context.Context) ([]models.Image, error) {
	rows, err := r.db.Query(ctx, "SELECT id, user_id, title, url, style, created_at FROM images")
	if err != nil {
		return nil, fmt.Errorf("failed to query images: %w", err)
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var img models.Image
		var createdAt pgtype.Timestamptz

		err := rows.Scan(&img.ID, &img.UserID, &img.Title, &img.URL, &img.Style, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}

		// Преобразуем в time.Time
		if createdAt.Valid {
			img.CreatedAt = createdAt.Time
		}

		images = append(images, img)
	}

	return images, nil
}

// GetByUserID возвращает все изображения, принадлежащие указанному пользователю
func (r *ImageRepository) GetByUserID(ctx context.Context, userID int) ([]models.Image, error) {
	query := `SELECT id, user_id, title, url, style, created_at FROM images WHERE user_id = $1 ORDER BY id`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса изображений: %w", err)
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var img models.Image
		var createdAt pgtype.Timestamptz
		err := rows.Scan(&img.ID, &img.UserID, &img.Title, &img.URL, &img.Style, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования изображения: %w", err)
		}
		// Преобразуем в time.Time
		if createdAt.Valid {
			img.CreatedAt = createdAt.Time
		}

		images = append(images, img)
	}

	return images, nil
}

func (r *ImageRepository) GetByIDAndUser(ctx context.Context, imageID string, userID int) (*models.Image, error) {
	var img models.Image
	var createdAt pgtype.Timestamptz

	query := `SELECT id, user_id, title, url, style, created_at FROM images WHERE id = $1 AND user_id = $2`
	err := r.db.QueryRow(ctx, query, imageID, userID).Scan(
		&img.ID, &img.UserID, &img.Title, &img.URL, &img.Style, &createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("ошибка при получении изображения: %w", err)
	}

	// Преобразуем в time.Time
	if createdAt.Valid {
		img.CreatedAt = createdAt.Time
	}

	return &img, nil
}

func (r *ImageRepository) Create(ctx context.Context, img *models.Image) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO images (id, user_id, title, url, style, created_at) VALUES($1, $2, $3, $4, $5, NOW())",
		img.ID, img.UserID, img.Title, img.URL, img.Style)
	return err
}
