//go:build wireinject
// +build wireinject

package app

import (
	"github.com/danielmesquitta/openfinance/internal/config"
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
	"github.com/google/wire"
)

func NewSyncAllUseCase() *usecase.SyncAll {
	wire.Build(
		validator.NewValidator,
		config.NewEnv,

		wire.Bind(new(companyapi.APIProvider), new(*brasilapi.Client)),
		brasilapi.NewClient,

		wire.Bind(new(gpt.Provider), new(*openai.OpenAIClient)),
		openai.NewOpenAIClient,

		wire.Bind(new(sheet.Provider), new(*notionapi.Client)),
		notionapi.NewClient,

		wire.Bind(new(openfinance.APIProvider), new(*pluggyapi.Client)),
		pluggyapi.NewClient,

		usecase.NewSyncOne,
		usecase.NewSyncAll,
	)

	return &usecase.SyncAll{}
}
