package restapi

import (
	"github.com/danielmesquitta/openfinance/internal/app/restapi/handler"
	"github.com/danielmesquitta/openfinance/internal/app/restapi/middleware"
	"github.com/danielmesquitta/openfinance/internal/app/restapi/router"
	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
	"github.com/danielmesquitta/openfinance/internal/provider/repo"
	"github.com/danielmesquitta/openfinance/internal/provider/repo/pgrepo"
	"github.com/danielmesquitta/openfinance/pkg/crypto"
	"github.com/danielmesquitta/openfinance/pkg/jwt"
	"github.com/danielmesquitta/openfinance/pkg/validator"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

func Start() {
	depsProvider := fx.Provide(
		// Config
		config.LoadEnv,

		// PKGs
		validator.NewValidator,
		jwt.NewIssuer,
		fx.Annotate(
			crypto.NewCrypto,
			fx.As(new(crypto.Encrypter)),
		),

		// Providers
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
		usecase.NewSyncAllUsersOpenFinanceDataToNotionUseCase,
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
		NewApp,
	)

	container := fx.New(
		depsProvider,
		fx.Invoke(func(*fiber.App) {}),
	)

	container.Run()
}
