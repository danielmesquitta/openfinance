package lambda

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/danielmesquitta/openfinance/internal/app"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase/ingest"
)

const (
	contentTypeHeader = "Content-Type"
	applicationJSON   = "application/json"
)

type LambdaHandler struct {
	ingestUseCase ingest.IngestExecutor
}

func NewLambdaHandler() (*LambdaHandler, error) {
	ingestUseCase, err := app.NewIngestUseCase()
	if err != nil {
		return nil, fmt.Errorf("initialize application: %w", err)
	}

	return newLambdaHandler(ingestUseCase), nil
}

func newLambdaHandler(ingestUseCase ingest.IngestExecutor) *LambdaHandler {
	return &LambdaHandler{ingestUseCase: ingestUseCase}
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

func (h *LambdaHandler) Handle(ctx context.Context) (Response, error) {
	startTime := time.Now()

	startDate, endDate := last7days()

	input := ingest.IngestInput{
		StartDate: startDate,
		EndDate:   endDate,
	}

	err := h.ingestUseCase.Execute(ctx, input)
	if err != nil {
		return newResponse(http.StatusInternalServerError, ErrorResponse{
			Error:   "ingest_failed",
			Message: fmt.Sprintf("Failed to execute ingest: %v", err),
		}), nil
	}

	duration := time.Since(startTime)

	return newResponse(http.StatusOK, SuccessResponse{
		Message:   "Ingest completed successfully",
		StartDate: startDate.Format(time.RFC3339),
		EndDate:   endDate.Format(time.RFC3339),
		Duration:  duration.String(),
	}), nil
}

func newResponse(statusCode int, value any) Response {
	body, err := json.Marshal(value)
	if err != nil {
		statusCode = http.StatusInternalServerError
		body = []byte(`{"error":"encoding_failed","message":"Failed to encode response"}`)
	}

	return Response{
		StatusCode: statusCode,
		Headers: map[string]string{
			contentTypeHeader: applicationJSON,
		},
		Body: string(body),
	}
}

func last7days() (startDate time.Time, endDate time.Time) {
	endDate = time.Now()
	startDate = endDate.AddDate(0, 0, -7)

	return startDate, endDate
}
