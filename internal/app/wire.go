//go:build wireinject
// +build wireinject

package app

import (
	"github.com/google/wire"

	"github.com/danielmesquitta/openfinance-to-sheets/internal/config"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/domain/entity"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/domain/usecase/ingest"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/pkg/validator"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/provider/companyapi"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/provider/companyapi/brasilapi"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/provider/gpt"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/provider/gpt/openai"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/provider/openfinance"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/provider/openfinance/pluggyapi"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/provider/sheet"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/provider/sheet/notionapi"
)

func ingestSettings(env *config.Env) entity.IngestSettings {
	return env.IngestSettings
}

func maxConcurrentOperations(env *config.Env) int {
	return env.MaxConcurrentOperations
}

func NewIngestUseCase() (*ingest.Ingest, error) {
	wire.Build(
		validator.NewValidator,
		config.NewEnv,
		ingestSettings,
		maxConcurrentOperations,

		wire.Bind(new(companyapi.APIProvider), new(*brasilapi.Client)),
		brasilapi.NewClient,

		wire.Bind(new(gpt.Provider), new(*openai.OpenAIClient)),
		openai.NewOpenAIClient,

		wire.Bind(new(sheet.Provider), new(*notionapi.Client)),
		notionapi.NewClient,

		wire.Bind(new(openfinance.APIProvider), new(*pluggyapi.Client)),
		pluggyapi.NewClient,

		ingest.NewIngest,
	)

	return nil, nil
}
