package http

import (
	"github.com/danielmesquitta/openfinance/config"
	"github.com/danielmesquitta/openfinance/internal/app/http/handler"
	"github.com/danielmesquitta/openfinance/internal/app/http/router"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
	"github.com/danielmesquitta/openfinance/internal/service/meupluggyapi"
	"github.com/danielmesquitta/openfinance/internal/service/notionapi"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

func Start() {
	depsProvider := fx.Provide(
		// Config
		config.LoadEnv,

		// Services
		meupluggyapi.NewClient,
		notionapi.NewClient,

		// Use cases
		usecase.NewOpenFinanceToNotionUseCase,

		// Handlers
		handler.NewOpenFinanceToNotionHandler,

		// Router
		router.NewRouter,

		// App
		newApp,
	)

	container := fx.New(
		depsProvider,
		fx.Invoke(func(*fiber.App) {}),
	)

	container.Run()
}
