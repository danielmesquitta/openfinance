package lambda

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/danielmesquitta/openfinance/internal/domain/usecase"
	"github.com/danielmesquitta/openfinance/internal/domain/usecase/mockusecase"
)

func TestLambdaHandlerSuccess(t *testing.T) {
	t.Parallel()

	syncUseCase := mockusecase.NewMockSyncExecutor(t)
	syncUseCase.EXPECT().
		Execute(mock.Anything, mock.MatchedBy(func(input usecase.SyncInput) bool {
			return input.EndDate.Sub(input.StartDate).Hours() == 7*24
		})).
		Return(nil).
		Once()

	response, err := newLambdaHandler(syncUseCase).Handle(t.Context())
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if response.Headers[contentTypeHeader] != applicationJSON {
		t.Fatalf("headers = %#v", response.Headers)
	}
	var body SuccessResponse
	if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Message != "Sync completed successfully" || body.StartDate == "" || body.EndDate == "" {
		t.Fatalf("body = %#v", body)
	}
}

func TestLambdaHandlerFailure(t *testing.T) {
	t.Parallel()

	syncUseCase := mockusecase.NewMockSyncExecutor(t)
	syncUseCase.EXPECT().Execute(mock.Anything, mock.Anything).Return(errors.New("failed")).Once()

	response, err := newLambdaHandler(syncUseCase).Handle(t.Context())
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d", response.StatusCode)
	}

	var body ErrorResponse
	if err := json.Unmarshal([]byte(response.Body), &body); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if body.Error != "sync_failed" {
		t.Fatalf("body = %#v", body)
	}
}

func TestNewResponseEncodingFallback(t *testing.T) {
	t.Parallel()

	response := newResponse(http.StatusOK, make(chan int))
	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("status = %d", response.StatusCode)
	}
	if response.Body != `{"error":"encoding_failed","message":"Failed to encode response"}` {
		t.Fatalf("body = %q", response.Body)
	}
}
