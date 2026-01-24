package repositories

import (
	"context"
	"fmt"
	"errors"
	"moduleExample/web-service-gin/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ImageRepository struct {
	db *pgxpool.Pool
}

func NewImageRepository(db *pgxpool.Pool) *ImageRepository {
	return &ImageRepository{db: db}
}

func (r *ImageRepository) GetAll(ctx context.Context) ([]models.Image, error) {
	rows, err := r.db.Query(ctx, "SELECT id, title, user_id, url, style FROM images")
	if err != nil {
		return nil, fmt.Errorf("failed to query images: %w", err)
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var img models.Image
		err := rows.Scan(&img.ID, &img.Title, &img.UserID, &img.URL, &img.Style)
		if err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images = append(images, img)
	}
	return images, nil
}

/*func (r *ImageRepository) GetByID(ctx context.Context, id string) (*models.Image, error) {
	var img models.Image
	err := r.db.QueryRow(ctx, "SELECT id, title, user_id, url, style FROM images WHERE id = $1", id).
		Scan(&img.ID, &img.Title, &img.UserID, &img.URL, &img.Style)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // не найдено
		}
		return nil, fmt.Errorf("failed to get image: %w", err)
	}
	return &img, nil
}
*/

// GetByUserID возвращает все изображения, принадлежащие указанному пользователю
func (r *ImageRepository) GetByUserID(ctx context.Context, userID int) ([]models.Image, error) {
	query := `SELECT id, user_id, title, url, style FROM images WHERE user_id = $1 ORDER BY id`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса изображений: %w", err)
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		var img models.Image
		err := rows.Scan(&img.ID, &img.UserID, &img.Title, &img.URL, &img.Style)
		if err != nil {
			return nil, fmt.Errorf("ошибка сканирования изображения: %w", err)
		}
		images = append(images, img)
	}

	return images, nil
}

func (r *ImageRepository) GetByIDAndUser(ctx context.Context, imageID string, userID int) (*models.Image, error) {
	var img models.Image
	query := `SELECT id, user_id, title, url, style FROM images WHERE id = $1 AND user_id = $2`
	err := r.db.QueryRow(ctx, query, imageID, userID).Scan(
		&img.ID, &img.UserID, &img.Title, &img.URL, &img.Style,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // изображение не найдено или не принадлежит пользователю
		}
		return nil, fmt.Errorf("ошибка при получении изображения: %w", err)
	}

	return &img, nil
}

func (r *ImageRepository) Create(ctx context.Context, img *models.Image) error {
	_, err := r.db.Exec(ctx,
		"INSERT INTO images (id, title, user_id, url, style) VALUES($1, $2, $3, $4, $5)",
		img.ID, img.Title, img.UserID, img.URL, img.Style)
	return err
}
