package usecase

import (
	"fmt"

	"github.com/danielmesquitta/openfinance/internal/provider/repo"
	"github.com/danielmesquitta/openfinance/pkg/jwt"
	"github.com/danielmesquitta/openfinance/pkg/validator"
	"github.com/jinzhu/copier"
)

type OAuthAuthenticationUseCase struct {
	userRepo  repo.UserRepo
	val       *validator.Validator
	jwtIssuer *jwt.JWTIssuer
}

func NewOAuthAuthenticationUseCase(
	userRepo repo.UserRepo,
	val *validator.Validator,
	jwtIssuer *jwt.JWTIssuer,
) *OAuthAuthenticationUseCase {
	return &OAuthAuthenticationUseCase{
		userRepo:  userRepo,
		val:       val,
		jwtIssuer: jwtIssuer,
	}
}

type OAuthAuthenticationDTO struct {
	Email string `json:"email" validate:"required,email"`
}

func (uc *OAuthAuthenticationUseCase) Execute(
	dto OAuthAuthenticationDTO,
) (accessToken string, expiresAt int64, err error) {
	if err := uc.val.Validate(dto); err != nil {
		return "", 0, err
	}

	user, err := uc.userRepo.GetUserByEmail(dto.Email)
	if err != nil {
		return "", 0, fmt.Errorf("error getting user by email: %w", err)
	}

	if userExists := user.ID != ""; userExists {
		return uc.jwtIssuer.NewAccessToken(user.ID)
	}

	if err := copier.Copy(&user, dto); err != nil {
		return "", 0, fmt.Errorf("error copying dto to user: %w", err)
	}

	if err := uc.userRepo.CreateUser(&user); err != nil {
		return "", 0, fmt.Errorf("error creating user: %w", err)
	}

	return uc.jwtIssuer.NewAccessToken(user.ID)
}
