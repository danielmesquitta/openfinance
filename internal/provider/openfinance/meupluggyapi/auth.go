package meupluggyapi

import (
	"encoding/json"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

type authResponse struct {
	APIKey string `json:"apiKey"`
}

type authRequest struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

func (c *Client) authenticate(
	clientID, clientSecret string,
) (string, error) {
	authRequest := authRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	res, err := c.client.R().SetBody(authRequest).Post("/auth")
	if err != nil {
		return "", entity.NewErr(err)
	}

	body := res.Body()
	if statusCode := res.StatusCode(); statusCode < 200 || statusCode >= 300 {
		return "", entity.NewErr(body)
	}

	data := authResponse{}
	if err := json.Unmarshal(res.Body(), &data); err != nil {
		return "", entity.NewErr(err)
	}

	if data.APIKey == "" {
		return "", entity.NewErr("api key is empty")
	}

	return data.APIKey, nil
}
