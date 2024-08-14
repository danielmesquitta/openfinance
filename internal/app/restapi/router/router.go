package router

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"

	"github.com/danielmesquitta/openfinance/internal/app/restapi/docs"
	"github.com/danielmesquitta/openfinance/internal/app/restapi/handler"
	"github.com/danielmesquitta/openfinance/internal/app/restapi/middleware"
	"github.com/danielmesquitta/openfinance/internal/config"
)

type Router struct {
	env                        *config.Env
	openFinanceToNotionHandler *handler.OpenFinanceToNotionHandler
	authHandler                *handler.AuthHandler
	settingHandler             *handler.SettingHandler
	middleware                 *middleware.Middleware
}

func NewRouter(
	env *config.Env,
	openFinanceToNotionHandler *handler.OpenFinanceToNotionHandler,
	authHandler *handler.AuthHandler,
	settingHandler *handler.SettingHandler,
	middleware *middleware.Middleware,
) *Router {
	return &Router{
		env:                        env,
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

	apiV1.Post(
		"/to-notion",
		r.middleware.EnsureBasicAuth,
		r.openFinanceToNotionHandler.SyncAllUsers,
	)

	apiV1.Post(
		"/users/me/settings",
		r.middleware.EnsureBearerAuth,
		r.settingHandler.Upsert,
	)

	docs.SwaggerInfo.BasePath = basePath
	app.Get("/docs/*", swagger.New())
}
