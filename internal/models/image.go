package models

import "time"

type Image struct {
	ID        string    `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	UserID    int       `json:"user_id" db:"user_id"`
	URL       string    `json:"url" db:"url"`
	Style     string    `json:"style" db:"style"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
