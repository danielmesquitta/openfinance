package meupluggyapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type AuthenticateResponse struct {
	APIKey string `json:"apiKey"`
}

func (c *Client) Authenticate() error {
	url := c.BaseURL.String() + "/auth"

	payload := strings.NewReader(fmt.Sprintf(
		"{\"clientId\":\"%s\",\"clientSecret\":\"%s\"}",
		c.ClientID,
		c.ClientSecret,
	))

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return parseResError(res)
	}

	decoder := json.NewDecoder(res.Body)

	data := AuthenticateResponse{}
	if err := decoder.Decode(&data); err != nil {
		return err
	}

	c.Token = data.APIKey

	return nil
}
