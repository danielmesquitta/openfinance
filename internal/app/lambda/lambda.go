package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/danielmesquitta/openfinance/internal/app"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
)

const (
	contentTypeHeader = "Content-Type"
	applicationJSON   = "application/json"
)

type LambdaHandler struct {
	syncAllUseCase *usecase.SyncAll
}

func NewLambdaHandler() *LambdaHandler {
	return &LambdaHandler{
		syncAllUseCase: app.NewSyncAllUseCase(),
	}
}

type Response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Metadata   map[string]any    `json:"metadata,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type SuccessResponse struct {
	Message   string `json:"message"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Duration  string `json:"duration"`
}

func (h *LambdaHandler) Handle(
	ctx context.Context,
	_ events.APIGatewayProxyRequest,
) (Response, error) {
	startTime := time.Now()

	startDateStr, endDateStr, err := h.executeSync(ctx)
	if err != nil {
		errorResponse := ErrorResponse{
			Error:   "sync_failed",
			Message: fmt.Sprintf("Failed to execute sync: %v", err),
		}

		body, err := json.Marshal(errorResponse)
		if err != nil {
			return Response{
				StatusCode: http.StatusInternalServerError,
				Headers: map[string]string{
					contentTypeHeader: applicationJSON,
				},
				Body: string(body),
			}, nil
		}

		return Response{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				contentTypeHeader: applicationJSON,
			},
			Body: string(body),
		}, nil
	}

	duration := time.Since(startTime)

	successResponse := SuccessResponse{
		Message:   "Sync completed successfully",
		StartDate: startDateStr,
		EndDate:   endDateStr,
		Duration:  duration.String(),
	}

	body, err := json.Marshal(successResponse)
	if err != nil {
		return Response{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				contentTypeHeader: applicationJSON,
			},
			Body: string(body),
		}, nil
	}

	return Response{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			contentTypeHeader: applicationJSON,
		},
		Body: string(body),
	}, nil
}

func (h *LambdaHandler) HandleScheduledEvent(
	ctx context.Context,
	_ events.CloudWatchEvent,
) error {
	_, _, err := h.executeSync(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute sync: %w", err)
	}

	return nil
}

func (h *LambdaHandler) executeSync(ctx context.Context) (startDateStr, endDateStr string, err error) {
	startDate, endDate := last15Days()

	startDateStr = startDate.Format(time.RFC3339)
	endDateStr = endDate.Format(time.RFC3339)

	syncDTO := usecase.SyncDTO{
		StartDate: startDateStr,
		EndDate:   endDateStr,
	}

	err = h.syncAllUseCase.Execute(ctx, syncDTO)
	if err != nil {
		return "", "", fmt.Errorf("failed to execute sync: %w", err)
	}

	return startDateStr, endDateStr, nil
}

func last15Days() (startDate time.Time, endDate time.Time) {
	endDate = time.Now()
	startDate = endDate.AddDate(0, 0, -15)

	return startDate, endDate
}
