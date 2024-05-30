package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"

	"github.com/danielmesquitta/openfinance/internal/app/http/docs"
	"github.com/danielmesquitta/openfinance/internal/app/http/handler"
)

type Router struct {
	openFinanceToNotionHandler *handler.OpenFinanceToNotionHandler
}

// @title OpenFinance to Notion API
// @version 1.0
// @description This API is responsible for syncing OpenFinance data to Notion.
// @contact.name Daniel Mesquita
// @contact.email danielmesquitta123@gmail.com
// @BasePath /
func NewRouter(
	openFinanceToNotionHandler *handler.OpenFinanceToNotionHandler,
) *Router {
	return &Router{
		openFinanceToNotionHandler: openFinanceToNotionHandler,
	}
}

func (r *Router) Register(
	app *fiber.App,
) {
	basePath := "/api/v1"

	apiV1 := app.Group(basePath)
	apiV1.Post(
		"/to-notion",
		r.openFinanceToNotionHandler.Sync,
	)

	docs.SwaggerInfo.BasePath = basePath
	app.Get("/docs/*", swagger.New())
}
