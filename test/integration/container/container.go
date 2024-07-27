package container

import (
	"sync"

	"github.com/danielmesquitta/openfinance/internal/app/http"
	"github.com/danielmesquitta/openfinance/internal/app/http/handler"
	"github.com/danielmesquitta/openfinance/internal/app/http/middleware"
	"github.com/danielmesquitta/openfinance/internal/app/http/router"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
	"github.com/danielmesquitta/openfinance/internal/provider/repo"
	"github.com/danielmesquitta/openfinance/internal/provider/repo/pgrepo"
	"github.com/danielmesquitta/openfinance/pkg/crypto"
	"github.com/danielmesquitta/openfinance/pkg/jwt"
	"github.com/danielmesquitta/openfinance/pkg/validator"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

var JwtIssuer *jwt.JWTIssuer
var Validator *validator.Validator

func NewApp(dbConnURL string) *fiber.App {
	env := loadTestEnv(dbConnURL)
	JwtIssuer = jwt.NewJWTIssuer(env)
	Validator = validator.NewValidator()

	var app *fiber.App
	wg := sync.WaitGroup{}
	wg.Add(1)

	depsProvider := []fx.Option{
		// Config
		fx.Supply(env),

		// PKGs
		fx.Supply(Validator),
		fx.Supply(JwtIssuer),

		fx.Provide(
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
			http.NewApp,
		),
		fx.Invoke(func(instance *fiber.App) {
			defer wg.Done()
			app = instance
		}),
	}

	container := fx.New(depsProvider...)

	go container.Run()

	wg.Wait()

	return app
}
