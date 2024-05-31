package handler

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
)

type OpenFinanceToNotionHandler struct {
	ouc *usecase.OpenFinanceToNotionUseCase
}

func NewOpenFinanceToNotionHandler(
	ouc *usecase.OpenFinanceToNotionUseCase,
) *OpenFinanceToNotionHandler {
	return &OpenFinanceToNotionHandler{
		ouc: ouc,
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
	if err := h.ouc.Execute(dto); err != nil {
		return fmt.Errorf(
			"error executing open finance to notion use case: %w",
			err,
		)
	}

	return c.SendStatus(http.StatusOK)
}
