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
	if statusCode := res.StatusCode(); statusCode < 200 || statusCode >= 300 {
		return "", fmt.Errorf("error response while authenticating: %+v", body)
	}

	data := authResponse{}
	if err := json.Unmarshal(res.Body(), &data); err != nil {
		return "", fmt.Errorf("failed to unmarshal while authenticating: %w", err)
	}

	if data.APIKey == "" {
		return "", errors.New("api key is empty")
	}

	return data.APIKey, nil
}
