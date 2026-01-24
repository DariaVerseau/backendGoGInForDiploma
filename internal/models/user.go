package models

type User struct {
	ID       int    `json:"id" db:"id"`
	Email    string `json:"email" db:"email"`
	Password string `json:"-" db:"password"` // не отправляем в JSON!
}
