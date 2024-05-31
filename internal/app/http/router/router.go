package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"

	"github.com/danielmesquitta/openfinance/internal/app/http/docs"
	"github.com/danielmesquitta/openfinance/internal/app/http/handler"
	"github.com/danielmesquitta/openfinance/internal/app/http/middleware"
)

type Router struct {
	openFinanceToNotionHandler *handler.OpenFinanceToNotionHandler
	authHandler                *handler.AuthHandler
	settingHandler             *handler.SettingHandler
	middleware                 *middleware.Middleware
}

func NewRouter(
	openFinanceToNotionHandler *handler.OpenFinanceToNotionHandler,
	authHandler *handler.AuthHandler,
	settingHandler *handler.SettingHandler,
	middleware *middleware.Middleware,
) *Router {
	return &Router{
		openFinanceToNotionHandler: openFinanceToNotionHandler,
		authHandler:                authHandler,
		settingHandler:             settingHandler,
		middleware:                 middleware,
	}
}

func (r *Router) Register(
	app *fiber.App,
) {
	basePath := "/api/v1"

	apiV1 := app.Group(basePath)

	apiV1.Get("/auth/login/:provider", r.authHandler.BeginOAuth)
	apiV1.Get("/auth/callback/:provider", r.authHandler.OAuthCallback)

	apiV1.Use(r.middleware.EnsureAuthenticated)

	apiV1.Post("/users/me/settings", r.settingHandler.Upsert)
	apiV1.Post(
		"/to-notion",
		r.openFinanceToNotionHandler.Sync,
	)

	docs.SwaggerInfo.BasePath = basePath
	app.Get("/docs/*", swagger.New())
}
