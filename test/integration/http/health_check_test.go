package http_test

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/danielmesquitta/openfinance/test/integration/container"
	"github.com/danielmesquitta/openfinance/test/integration/container/pgcontainer"
)

func TestHTTPHealthCheck(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		route        string
		expectedCode int
	}{
		{
			name:         "should successfully do liveness check",
			method:       "GET",
			route:        "/live",
			expectedCode: 200,
		},
		{
			name:         "should  successfully do readiness check",
			method:       "GET",
			route:        "/ready",
			expectedCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dbConnURL, terminate := pgcontainer.NewPgContainer(
				context.Background(),
			)
			defer terminate()

			app := container.NewApp(dbConnURL)
			defer func() {
				err := app.Shutdown()
				if err != nil {
					t.Fatalf("failed to shutdown app: %s", err)
				}
			}()

			req := httptest.NewRequest(tt.method, tt.route, nil)

			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("failed to make request: %s", err)
			}

			if resp.StatusCode != tt.expectedCode {
				t.Errorf(
					"expected status code %d, got %d",
					tt.expectedCode,
					resp.StatusCode,
				)
			}
		})
	}
}
