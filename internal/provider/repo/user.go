package repo

import "github.com/danielmesquitta/openfinance/internal/domain/entity"

type UserRepo interface {
	GetUserByID(id string) (entity.User, error)
	GetUserWithSettingByID(id string) (entity.User, error)
	GetUserByEmail(email string) (entity.User, error)
	CreateUser(user *entity.User) error
}
