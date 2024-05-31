package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/danielmesquitta/openfinance/internal/app/http/middleware"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
)

type SettingHandler struct {
	uuc *usecase.UpsertUserSettingUseCase
}

func NewSettingHandler(
	uuc *usecase.UpsertUserSettingUseCase,
) *SettingHandler {
	return &SettingHandler{
		uuc: uuc,
	}
}

// @Summary Upsert user setting.
// @Description This endpoint is responsible for updating and creating user settings.
// @Tags Setting
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body usecase.UpsertUserSettingDTO true "Request body"
// @Success 200
// @Failure 400 {object} dto.ErrorResponseDTO
// @Failure 500 {object} dto.ErrorResponseDTO
// @Router /users/me/settings [post]
func (h *SettingHandler) Upsert(c *fiber.Ctx) error {
	dto := usecase.UpsertUserSettingDTO{}
	if err := c.BodyParser(&dto); err != nil {
		return err
	}

	userID := c.Locals(middleware.UserIDKey).(string)
	dto.UserID = userID

	if err := h.uuc.Execute(dto); err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}
