package http

import (
	"context"

	"github.com/danielmesquitta/openfinance/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/fx"
)

func newApp(
	lc fx.Lifecycle,
	env *config.Env,
) *fiber.App {
	app := fiber.New()

	app.Use(recover.New())

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
