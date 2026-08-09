//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
	"github.com/danielmesquitta/openfinance/internal/pkg/validator"
	"github.com/danielmesquitta/openfinance/internal/provider/companyapi"
	"github.com/danielmesquitta/openfinance/internal/provider/companyapi/brasilapi"
	"github.com/danielmesquitta/openfinance/internal/provider/gpt"
	"github.com/danielmesquitta/openfinance/internal/provider/gpt/openai"
	"github.com/danielmesquitta/openfinance/internal/provider/openfinance"
	"github.com/danielmesquitta/openfinance/internal/provider/openfinance/pluggyapi"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet/notionapi"
)

func syncSettings(env *config.Env) entity.SyncSettings {
	return env.SyncSettings
}

func maxConcurrentOperations(env *config.Env) int {
	return env.MaxConcurrentOperations
}

func NewSyncUseCase() (*usecase.Sync, error) {
	wire.Build(
		validator.NewValidator,
		config.NewEnv,
		syncSettings,
		maxConcurrentOperations,

		wire.Bind(new(companyapi.APIProvider), new(*brasilapi.Client)),
		brasilapi.NewClient,

		wire.Bind(new(gpt.Provider), new(*openai.OpenAIClient)),
		openai.NewOpenAIClient,

		wire.Bind(new(sheet.Provider), new(*notionapi.Client)),
		notionapi.NewClient,

		wire.Bind(new(openfinance.APIProvider), new(*pluggyapi.Client)),
		pluggyapi.NewClient,

		usecase.NewSync,
	)

	return nil, nil
}
