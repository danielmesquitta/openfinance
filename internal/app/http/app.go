package http

import (
	"context"

	"github.com/danielmesquitta/openfinance/internal/app/http/middleware"
	"github.com/danielmesquitta/openfinance/internal/app/http/router"
	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/pkg/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/idempotency"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/google"
	"github.com/shareed2k/goth_fiber"
	"go.uber.org/fx"
)

func newApp(
	lc fx.Lifecycle,
	env *config.Env,
	middleware *middleware.Middleware,
	log *logger.Logger,
	router *router.Router,
) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(idempotency.New())

	goth_fiber.SessionStore = session.New(session.Config{
		CookiePath:     "/",
		CookieHTTPOnly: true,
		CookieSecure:   env.Environment != config.DevelopmentEnv,
	})

	goth.UseProviders(
		google.New(
			env.GoogleOAUTHClientID,
			env.GoogleOAUTHClientSecret,
			env.ApiURL+"/api/v1/auth/callback/google",
		),
	)

	router.Register(app)

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				err := app.Listen(":" + env.Port)

				if err != nil {
					panic(err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			return app.Shutdown()
		},
	})

	return app
}
