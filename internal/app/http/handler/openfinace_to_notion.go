package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/danielmesquitta/openfinance/internal/app/http/dto"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
)

type OpenFinanceToNotionHandler struct {
	uc *usecase.OpenFinanceToNotionUseCase
}

func NewOpenFinanceToNotionHandler(
	uc *usecase.OpenFinanceToNotionUseCase,
) *OpenFinanceToNotionHandler {
	return &OpenFinanceToNotionHandler{
		uc: uc,
	}
}

// @Summary OpenFinance to Notion.
// @Description This endpoint is responsible for syncing OpenFinance data to Notion.
// @Tags Notion
// @Accept json
// @Produce json
// @Success 200
// @Failure 500 {object} dto.ErrorResponseDTO
// @Router /to-notion [post]
func (h *OpenFinanceToNotionHandler) Sync(c *fiber.Ctx) error {
	if err := h.uc.Execute(); err != nil {
		return c.Status(fiber.StatusInternalServerError).
			JSON(dto.ErrorResponseDTO{Message: err.Error()})
	}

	return c.SendStatus(fiber.StatusOK)
}
