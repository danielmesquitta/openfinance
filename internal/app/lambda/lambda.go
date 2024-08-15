package lambda

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
	"github.com/danielmesquitta/openfinance/internal/pkg/validator"
	"github.com/danielmesquitta/openfinance/internal/provider/companyapi/brasilapi"
	"github.com/danielmesquitta/openfinance/internal/provider/gpt/openai"
	"github.com/danielmesquitta/openfinance/internal/provider/openfinance/meupluggyapi"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet/notionapi"
)

func Handler(
	_ events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	val := validator.NewValidator()
	env := config.NewEnv(val)
	companyAPIProvider := brasilapi.NewClient()
	gptProvider := openai.NewOpenAIClient(env)
	sheetProvider := notionapi.NewClient(env)
	openFinanceAPIProvider := meupluggyapi.NewClient(env)

	syncOneUseCase := usecase.NewSyncOne(
		val,
		companyAPIProvider,
		gptProvider,
		sheetProvider,
		openFinanceAPIProvider,
	)

	syncAllUseCase := usecase.NewSyncAll(
		val,
		env,
		syncOneUseCase,
	)

	err := syncAllUseCase.Execute(usecase.SyncDTO{})
	if err != nil {
		slog.Error(
			err.Error(),
			"error", err,
		)
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       fmt.Sprintf("{\"message\": \"%s\"}", err.Error()),
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}
