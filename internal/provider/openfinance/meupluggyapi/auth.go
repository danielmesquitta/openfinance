package meupluggyapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type AuthenticateResponse struct {
	APIKey string `json:"apiKey"`
}

func authenticate(
	baseURL url.URL,
	clientID, clientSecret string,
) (string, error) {
	url := baseURL.String() + "/auth"

	payload := strings.NewReader(fmt.Sprintf(
		"{\"clientId\":\"%s\",\"clientSecret\":\"%s\"}",
		clientID,
		clientSecret,
	))

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %w", err)
	}
	if res == nil {
		return "", fmt.Errorf("response is nil")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", parseResError(res)
	}

	decoder := json.NewDecoder(res.Body)

	data := AuthenticateResponse{}
	if err := decoder.Decode(&data); err != nil {
		return "", fmt.Errorf("error decoding response: %w", err)
	}

	return data.APIKey, nil
}
