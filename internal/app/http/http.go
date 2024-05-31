package http

import (
	"github.com/danielmesquitta/openfinance/internal/app/http/handler"
	"github.com/danielmesquitta/openfinance/internal/app/http/middleware"
	"github.com/danielmesquitta/openfinance/internal/app/http/router"
	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
	"github.com/danielmesquitta/openfinance/internal/provider/openfinance/meupluggyapi"
	"github.com/danielmesquitta/openfinance/internal/provider/repo"
	"github.com/danielmesquitta/openfinance/internal/provider/repo/pgrepo"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet/notionapi"
	"github.com/danielmesquitta/openfinance/pkg/jwt"
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
		jwt.NewJWTIssuer,

		// Providers
		meupluggyapi.NewClient,
		notionapi.NewClient,
		pgrepo.NewPgDBConn,
		fx.Annotate(
			pgrepo.NewUserPgRepo,
			fx.As(new(repo.UserRepo)),
		),
		fx.Annotate(
			pgrepo.NewSettingPgRepo,
			fx.As(new(repo.SettingRepo)),
		),

		// Use cases
		usecase.NewOpenFinanceToNotionUseCase,
		usecase.NewOAuthAuthenticationUseCase,
		usecase.NewUpsertUserSettingUseCase,

		// Middleware
		middleware.NewMiddleware,

		// Handlers
		handler.NewOpenFinanceToNotionHandler,
		handler.NewAuthHandler,
		handler.NewSettingHandler,

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
