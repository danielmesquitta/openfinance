package http

import (
	"github.com/danielmesquitta/openfinance/config"
	"github.com/danielmesquitta/openfinance/internal/app/http/handler"
	"github.com/danielmesquitta/openfinance/internal/app/http/middleware"
	"github.com/danielmesquitta/openfinance/internal/app/http/router"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
	"github.com/danielmesquitta/openfinance/internal/provider/openfinance/meupluggyapi"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet/notionapi"
	"github.com/danielmesquitta/openfinance/pkg/logger"
	"github.com/danielmesquitta/openfinance/pkg/validator"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

func Start() {
	depsProvider := fx.Provide(
		// Config
		config.LoadEnv,

		// PKGs
		logger.NewLogger,
		validator.NewValidator,

		// Services
		meupluggyapi.NewClient,
		notionapi.NewClient,

		// Use cases
		usecase.NewOpenFinanceToNotionUseCase,

		// Middleware
		middleware.NewMiddleware,

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
