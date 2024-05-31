package handler

import (
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

// @Summary OAuth Callback.
// @Description This endpoint is responsible for receiving oauth callbacks.
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body usecase.OAuthAuthenticationDTO true "Auth"
// @Success 200
// @Failure 400 {object} dto.ErrorResponseDTO
// @Failure 500 {object} dto.ErrorResponseDTO
// @Router /users [post]
func (h *AuthHandler) OAuthCallback(c *fiber.Ctx) error {
	user, err := goth_fiber.CompleteUserAuth(c)
	if err != nil {
		return err
	}

	data := usecase.OAuthAuthenticationDTO{
		Email: user.Email,
	}
	accessToken, expiresAt, err := h.ouc.Execute(data)
	if err != nil {
		return err
	}

	return c.Status(http.StatusOK).JSON(dto.OauthCallbackResponseDTO{
		AccessToken: accessToken,
		ExpiresAt:   expiresAt,
	})
}
