package cli

import (
	"log/slog"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
	"github.com/danielmesquitta/openfinance/internal/pkg/validator"
	"github.com/danielmesquitta/openfinance/internal/provider/companyapi/brasilapi"
	"github.com/danielmesquitta/openfinance/internal/provider/gpt/openai"
	"github.com/danielmesquitta/openfinance/internal/provider/openfinance/meupluggyapi"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet/notionapi"
)

func main() {
	val := validator.NewValidator()
	env := config.NewEnv(val)
	companyAPIProvider := brasilapi.NewClient()
	gptProvider := openai.NewOpenAIClient(env)
	sheetProvider := notionapi.NewClient(env)
	openFinanceAPIProvider := meupluggyapi.NewClient(env)

	u := usecase.NewSyncAll(
		val,
		env,
		companyAPIProvider,
		gptProvider,
		sheetProvider,
		openFinanceAPIProvider,
	)

	err := u.Execute(usecase.SyncAllDTO{})
	if err != nil {
		panic(err)
	}

	slog.Info(
		"SyncAll executed successfully",
	)
}
