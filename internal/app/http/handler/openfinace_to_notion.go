package handler

import (
	"net/http"

	"github.com/gofiber/fiber/v2"

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
// @Param start_date query string false "Start date (format RFC3339: 2006-01-02T15:04:05Z07:00)"
// @Param end_date query string false "End date (format RFC3339: 2006-01-02T15:04:05Z07:00)"
// @Success 200
// @Failure 400 {object} dto.ErrorResponseDTO
// @Failure 500 {object} dto.ErrorResponseDTO
// @Router /to-notion [post]
func (h *OpenFinanceToNotionHandler) Sync(c *fiber.Ctx) error {
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")

	dto := usecase.OpenFinanceToNotionUseCaseDTO{
		StartDate: startDate,
		EndDate:   endDate,
	}
	if err := h.uc.Execute(dto); err != nil {
		return err
	}

	return c.SendStatus(http.StatusOK)
}
