package handler

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"

	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
)

type OpenFinanceToNotionHandler struct {
	uc *usecase.SyncAllUsersOpenFinanceDataToNotionUseCase
}

func NewOpenFinanceToNotionHandler(
	uc *usecase.SyncAllUsersOpenFinanceDataToNotionUseCase,
) *OpenFinanceToNotionHandler {
	return &OpenFinanceToNotionHandler{
		uc: uc,
	}
}

// @Summary Sync all users OpenFinance data to Notion.
// @Description This endpoint is responsible for syncing all users OpenFinance data to Notion.
// @Tags Notion
// @Accept json
// @Produce json
// @Param start_date query string false "Start date (format RFC3339: 2006-01-02T15:04:05Z07:00)"
// @Param end_date query string false "End date (format RFC3339: 2006-01-02T15:04:05Z07:00)"
// @Success 200
// @Failure 400 {object} dto.ErrorResponseDTO
// @Failure 500 {object} dto.ErrorResponseDTO
// @Router /to-notion [post]
func (h *OpenFinanceToNotionHandler) SyncAllUsers(c *fiber.Ctx) error {
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")

	dto := usecase.SyncAllUsersOpenFinanceDataToNotionDTO{
		StartDate: startDate,
		EndDate:   endDate,
	}
	if err := h.uc.Execute(dto); err != nil {
		return fmt.Errorf(
			"error executing open finance to notion use case: %w",
			err,
		)
	}

	return c.SendStatus(http.StatusOK)
}
