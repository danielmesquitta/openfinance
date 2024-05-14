package router

import (
	"github.com/danielmesquitta/openfinance/internal/app/http/handler"
	"github.com/gofiber/fiber/v2"
)

type Router struct {
	app                        *fiber.App
	openFinanceToNotionHandler *handler.OpenFinanceToNotionHandler
}

func NewRouter(
	app *fiber.App,
	openFinanceToNotionHandler *handler.OpenFinanceToNotionHandler,
) *Router {
	basePath := "/api/v1"

	apiV1 := app.Group(basePath)
	apiV1.Get(
		"/to-notion",
		openFinanceToNotionHandler.Do,
	)

	return &Router{
		app:                        app,
		openFinanceToNotionHandler: openFinanceToNotionHandler,
	}
}
