package pluggyapi

import (
	"encoding/json"

	"github.com/danielmesquitta/openfinance/internal/domain/errs"
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
		return "", errs.New(err)
	}

	body := res.Body()
	if statusCode := res.StatusCode(); statusCode < 200 || statusCode >= 300 {
		return "", errs.New(body)
	}

	data := authResponse{}
	if err := json.Unmarshal(res.Body(), &data); err != nil {
		return "", errs.New(err)
	}

	if data.APIKey == "" {
		return "", errs.New("api key is empty")
	}

	return data.APIKey, nil
}
