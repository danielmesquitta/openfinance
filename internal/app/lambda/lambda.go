package lambda

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/danielmesquitta/openfinance/internal/app"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
)

func Handler(
	_ events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	syncAllUseCase := app.NewSyncAllUseCase()

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
