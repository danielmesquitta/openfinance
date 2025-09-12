package pluggyapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

type authResponse struct {
	APIKey string `json:"apiKey"`
}

type authRequest struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

func (c *Client) authenticate(
	ctx context.Context,
	clientID, clientSecret string,
) (string, error) {
	authRequest := authRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	res, err := c.client.R().SetContext(ctx).SetBody(authRequest).Post("/auth")
	if err != nil {
		return "", fmt.Errorf("failed to authenticate: %w", err)
	}

	body := res.Body()
	if res.IsError() {
		return "", fmt.Errorf("error response while authenticating: %s", body)
	}

	data := authResponse{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal while authenticating: %w", err)
	}

	if data.APIKey == "" {
		return "", errors.New("api key is empty")
	}

	return data.APIKey, nil
}
