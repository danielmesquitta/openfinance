package repo

import "github.com/danielmesquitta/openfinance/internal/domain/entity"

type CreateUserDTO struct {
	Email string `json:"email"`
}

type UserRepo interface {
	GetFullUserByID(id string) (entity.User, error)
	GetUserByEmail(email string) (entity.User, error)
	CreateUser(dto CreateUserDTO) (entity.User, error)
}
