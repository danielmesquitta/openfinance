package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/danielmesquitta/openfinance/internal/app/restapi/dto"
	"github.com/danielmesquitta/openfinance/test/integration/container"
	"github.com/danielmesquitta/openfinance/test/integration/container/pgcontainer"
)

func TestHTTPUpsertUserSetting(t *testing.T) {
	tests := []struct {
		name         string
		method       string
		route        string
		body         dto.UpsertUserSettingRequestDTO
		userID       string
		expectedCode int
	}{
		{
			name:   "should create user setting",
			method: "POST",
			route:  "/api/v1/users/me/settings",
			body: dto.UpsertUserSettingRequestDTO{
				NotionToken:           "new_notion_token",
				NotionPageID:          "new_notion_page_id",
				MeuPluggyClientID:     "new_meu_pluggy_client_id",
				MeuPluggyClientSecret: "new_meu_pluggy_client_secret",
				MeuPluggyAccountIDs:   []string{"new_meu_pluggy_account_id"},
			},
			userID:       pgcontainer.TestUserWithoutSetting.ID,
			expectedCode: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbConnURL, terminate := pgcontainer.NewPgContainer(
				context.Background(),
				pgcontainer.WithSeeds(
					pgcontainer.SeedTestUserWithoutSetting,
					pgcontainer.SeedTestUserWithSetting,
				),
			)
			defer terminate()

			app := container.NewApp(dbConnURL)
			defer func() {
				err := app.Shutdown()
				if err != nil {
					t.Fatalf("failed to shutdown app: %s", err)
				}
			}()

			jsonBody, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatalf("failed to marshal body: %s", err)
			}

			bytesBody := bytes.NewReader(jsonBody)

			req := httptest.NewRequest(tt.method, tt.route, bytesBody)

			accessToken, _, err := container.Issuer.NewAccessToken(tt.userID)
			if err != nil {
				t.Fatalf("failed to create access token: %s", err)
			}

			req.Header.Set("Authorization", "Bearer "+accessToken)
			req.Header.Set("Content-Type", "application/json")

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
