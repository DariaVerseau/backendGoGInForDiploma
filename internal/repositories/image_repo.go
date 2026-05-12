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

func scanImage(row pgx.Rows) (*models.Image, error) {
	var img models.Image
	var createdAt pgtype.Timestamptz
	var style pgtype.Text // ← используем pgtype.Text для поддержки NULL

	err := row.Scan(&img.ID, &img.UserID, &img.Title, &img.URL, &style, &createdAt)
	if err != nil {
		return nil, err
	}

	// Обрабатываем style (может быть NULL)
	if style.Valid {
		img.Style = style.String
	} else {
		img.Style = "" // Значение по умолчанию
	}

	if createdAt.Valid {
		img.CreatedAt = createdAt.Time
	}
	return &img, nil
}

func (r *ImageRepository) GetAll(ctx context.Context) ([]models.Image, error) {
	rows, err := r.db.Query(ctx, "SELECT id, user_id, title, url, style, created_at FROM images ORDER BY created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query images: %w", err)
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		img, err := scanImage(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images = append(images, *img)
	}
	return images, nil
}

func (r *ImageRepository) GetByUserID(ctx context.Context, userID int) ([]models.Image, error) {
	query := `SELECT id, user_id, title, url, style, created_at FROM images WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query images by user: %w", err)
	}
	defer rows.Close()

	var images []models.Image
	for rows.Next() {
		img, err := scanImage(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan image: %w", err)
		}
		images = append(images, *img)
	}
	return images, nil
}

func (r *ImageRepository) GetByIDAndUser(ctx context.Context, imageID string, userID int) (*models.Image, error) {
	var img models.Image
	var createdAt pgtype.Timestamptz
	var style pgtype.Text

	query := `SELECT id, user_id, title, url, style, created_at FROM images WHERE id = $1 AND user_id = $2`
	err := r.db.QueryRow(ctx, query, imageID, userID).Scan(
		&img.ID, &img.UserID, &img.Title, &img.URL, &style, &createdAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get image: %w", err)
	}

	if style.Valid {
		img.Style = style.String
	}
	if createdAt.Valid {
		img.CreatedAt = createdAt.Time
	}
	return &img, nil
}

func (r *ImageRepository) Create(ctx context.Context, img *models.Image) error {
	query := `INSERT INTO images (id, user_id, title, url, style, created_at) 
	          VALUES ($1, $2, $3, $4, $5, NOW())`

	// Если style пустой, вставляем NULL
	var style interface{}
	if img.Style == "" {
		style = nil
	} else {
		style = img.Style
	}

	_, err := r.db.Exec(ctx, query, img.ID, img.UserID, img.Title, img.URL, style)
	if err != nil {
		return fmt.Errorf("failed to create image: %w", err)
	}
	return nil
}
