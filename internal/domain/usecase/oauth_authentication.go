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

	params := repo.CreateUserDTO{}
	if err := copier.Copy(&params, dto); err != nil {
		return "", 0, fmt.Errorf("error copying dto to params: %w", err)
	}

	createdUser, err := uc.userRepo.CreateUser(params)
	if err != nil {
		return "", 0, fmt.Errorf("error creating user: %w", err)
	}

	return uc.jwtIssuer.NewAccessToken(createdUser.ID)
}
