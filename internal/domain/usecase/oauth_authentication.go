package usecase

import (
	"fmt"

	"github.com/danielmesquitta/openfinance/internal/provider/repo"
	"github.com/danielmesquitta/openfinance/pkg/jwt"
	"github.com/danielmesquitta/openfinance/pkg/validator"
	"github.com/jinzhu/copier"
)

type OAuthAuthenticationUseCase struct {
	ur repo.UserRepo
	v  *validator.Validator
	j  *jwt.JWTIssuer
}

func NewOAuthAuthenticationUseCase(
	ur repo.UserRepo,
	v *validator.Validator,
	j *jwt.JWTIssuer,
) *OAuthAuthenticationUseCase {
	return &OAuthAuthenticationUseCase{
		ur: ur,
		v:  v,
		j:  j,
	}
}

type OAuthAuthenticationDTO struct {
	Email string `json:"email" validate:"required,email"`
}

func (uc *OAuthAuthenticationUseCase) Execute(
	dto OAuthAuthenticationDTO,
) (accessToken string, expiresAt int64, err error) {
	if err := uc.v.Validate(dto); err != nil {
		return "", 0, err
	}

	user, err := uc.ur.GetUserByEmail(dto.Email)
	if err != nil {
		return "", 0, fmt.Errorf("error getting user by email: %w", err)
	}

	if userExists := user.ID != ""; userExists {
		return uc.j.NewAccessToken(user.ID)
	}

	if err := copier.Copy(&user, dto); err != nil {
		return "", 0, fmt.Errorf("error copying dto to user: %w", err)
	}

	if err := uc.ur.CreateUser(&user); err != nil {
		return "", 0, fmt.Errorf("error creating user: %w", err)
	}

	return uc.j.NewAccessToken(user.ID)
}
