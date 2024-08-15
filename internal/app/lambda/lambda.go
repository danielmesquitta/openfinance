package lambda

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
	"github.com/danielmesquitta/openfinance/internal/provider/companyapi/brasilapi"
	"github.com/danielmesquitta/openfinance/internal/provider/gpt/openai"
	"github.com/danielmesquitta/openfinance/internal/provider/repo/jsonrepo"
	"github.com/danielmesquitta/openfinance/pkg/crypto"
	"github.com/danielmesquitta/openfinance/pkg/validator"
)

func Handler() (events.APIGatewayProxyResponse, error) {
	val := validator.NewValidator()
	env := config.LoadEnv(val)
	cry := crypto.NewCrypto(env)
	settingRepo := jsonrepo.NewSettingJSONRepo(cry)
	companyAPI := brasilapi.NewClient()
	gptProvider := openai.NewOpenAIClient(env)

	u := usecase.NewSyncAllUsersOpenFinanceDataToNotionUseCase(
		val,
		cry,
		settingRepo,
		companyAPI,
		gptProvider,
	)

	err := u.Execute(usecase.SyncAllUsersOpenFinanceDataToNotionDTO{})

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
