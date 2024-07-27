package handler

import (
	"fmt"
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
// @Param request body dto.UpsertUserSettingRequestDTO true "Request body"
// @Success 200 {object} entity.Setting
// @Failure 400 {object} dto.ErrorResponseDTO
// @Failure 500 {object} dto.ErrorResponseDTO
// @Router /users/me/settings [post]
func (h *SettingHandler) Upsert(c *fiber.Ctx) error {
	data := usecase.UpsertUserSettingDTO{}
	if err := c.BodyParser(&data); err != nil {
		return fmt.Errorf("error parsing request body: %w", err)
	}

	userID := c.Locals(middleware.UserIDKey).(string)
	data.UserID = userID

	setting, err := h.uuc.Execute(data)
	if err != nil {
		return fmt.Errorf(
			"error executing upsert user setting use case: %w",
			err,
		)
	}

	return c.Status(http.StatusOK).JSON(setting)
}
