package handler

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/shareed2k/goth_fiber"

	"github.com/danielmesquitta/openfinance/internal/app/http/dto"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
)

type AuthHandler struct {
	ouc *usecase.OAuthAuthenticationUseCase
}

func NewAuthHandler(
	ouc *usecase.OAuthAuthenticationUseCase,
) *AuthHandler {
	return &AuthHandler{
		ouc: ouc,
	}
}

// @Summary BeginOAuth.
// @Description This endpoint is responsible for starting OAuth authentication.
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200
// @Failure 500 {object} dto.ErrorResponseDTO
// @Router /auth/login/google [get]
func (h *AuthHandler) BeginOAuth(c *fiber.Ctx) error {
	return goth_fiber.BeginAuthHandler(c)
}

func (h *AuthHandler) OAuthCallback(c *fiber.Ctx) error {
	user, err := goth_fiber.CompleteUserAuth(c)
	if err != nil {
		return fmt.Errorf("error completing user auth: %w", err)
	}

	data := usecase.OAuthAuthenticationDTO{
		Email: user.Email,
	}
	accessToken, expiresAt, err := h.ouc.Execute(data)
	if err != nil {
		return fmt.Errorf(
			"error executing oauth authentication use case: %w",
			err,
		)
	}

	return c.Status(http.StatusOK).JSON(dto.OauthCallbackResponseDTO{
		AccessToken: accessToken,
		ExpiresAt:   expiresAt,
	})
}
