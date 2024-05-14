package handler

import (
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

func (h *OpenFinanceToNotionHandler) Do(c *fiber.Ctx) error {
	if err := h.uc.Execute(); err != nil {
		return c.Status(fiber.StatusInternalServerError).
			SendString(err.Error())
	}

	return c.JSON(map[string]bool{"ok": true})
}
