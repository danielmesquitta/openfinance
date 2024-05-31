package entity

import (
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	UpdatedAt time.Time `json:"updated_at"`
	Setting   *Setting  `json:"setting"`
}
