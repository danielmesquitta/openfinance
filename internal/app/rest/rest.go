package rest

import (
	"github.com/gofiber/fiber/v2"

	"github.com/danielmesquitta/openfinance/config"
	"github.com/danielmesquitta/openfinance/internal/app/rest/handler"
	"github.com/danielmesquitta/openfinance/internal/service/meupluggyapi"
	"github.com/danielmesquitta/openfinance/internal/service/notionapi"
)

type AccountType int

const (
	BankAccount AccountType = iota
	CreditCard
)

func Start() error {
	e := config.LoadEnv()

	meupluggyAPIClient := meupluggyapi.NewClient(
		e.MeuPluggyClientID,
		e.MeuPluggyClientSecret,
	)

	notionAPIClient := notionapi.NewClient(e.NotionToken)

	app := fiber.New()

	app.Post(
		"/",
		handler.OpenFinanceToNotionHandler(
			e.NotionPageID,
			e.MeuPluggyAccountIDs,
			meupluggyAPIClient,
			notionAPIClient,
		),
	)

	app.Listen(":" + e.Port)

	return nil
}
