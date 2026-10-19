package models

import (
	"time"

	"github.com/google/uuid"
)

type Image struct {
	ID         uuid.UUID `json:"id" db:"id"`
	Title      string    `json:"title" db:"title"`
	UserID     int64     `json:"user_id" db:"user_id"`
	StorageKey string    `json:"storage_key" db:"storage_key"`
	Style      string    `json:"style" db:"style"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
