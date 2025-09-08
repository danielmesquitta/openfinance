package lambda

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/danielmesquitta/openfinance/internal/app"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
)

func Handler(
	_ events.APIGatewayProxyRequest,
) (events.APIGatewayProxyResponse, error) {
	syncAllUseCase := app.NewSyncAllUseCase()

	ctx := context.Background()

	startOfLastMonth := time.Date(
		time.Now().Year(),
		time.Now().Month()-1,
		1,
		0,
		0,
		0,
		0,
		time.Local,
	)
	endOfLastMonth := startOfLastMonth.AddDate(
		0,
		1,
		-1,
	)

	err := syncAllUseCase.Execute(ctx, usecase.SyncDTO{
		StartDate: startOfLastMonth.Format(time.RFC3339),
		EndDate:   endOfLastMonth.Format(time.RFC3339),
	})
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
		StatusCode: http.StatusOK,
		Headers:    map[string]string{"Content-Type": "application/json"},
	}, nil
}
