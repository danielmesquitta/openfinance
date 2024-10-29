package meupluggyapi

import (
	"bytes"
	"encoding/json"
	"net/http"

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
	url := c.baseURL
	url.Path = "/auth"

	authRequest := authRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	authRequestBytes, err := json.Marshal(authRequest)
	if err != nil {
		return "", entity.NewErr(err)
	}

	payload := bytes.NewReader(authRequestBytes)

	req, err := http.NewRequest("POST", url.String(), payload)
	if err != nil {
		return "", entity.NewErr(err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", entity.NewErr(err)
	}
	if res == nil {
		return "", entity.NewErr("response is nil")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", parseResError(res)
	}

	decoder := json.NewDecoder(res.Body)

	data := authResponse{}
	if err := decoder.Decode(&data); err != nil {
		return "", entity.NewErr(err)
	}

	return data.APIKey, nil
}
