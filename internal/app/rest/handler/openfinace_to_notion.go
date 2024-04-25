package handler

import (
	"github.com/gofiber/fiber/v2"

	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
	"github.com/danielmesquitta/openfinance/internal/service/meupluggyapi"
	"github.com/danielmesquitta/openfinance/internal/service/notionapi"
)

func OpenFinanceToNotionHandler(
	notionPageID string,
	meuPluggyAccountIDs []string,
	meupluggyAPIClient *meupluggyapi.Client,
	notionAPIClient *notionapi.Client,
) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {
		if err := usecase.OpenFinanceToNotionUseCase(
			notionPageID,
			meuPluggyAccountIDs,
			meupluggyAPIClient,
			notionAPIClient,
		); err != nil {
			return c.Status(fiber.StatusInternalServerError).
				SendString(err.Error())
		}

		return c.JSON(map[string]bool{"ok": true})
	}
}
