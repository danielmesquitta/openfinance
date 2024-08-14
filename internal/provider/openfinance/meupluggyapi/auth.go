package meupluggyapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type AuthenticateResponse struct {
	APIKey string `json:"apiKey"`
}

type AuthenticateRequest struct {
	ClientID     string `json:"clientId"`
	ClientSecret string `json:"clientSecret"`
}

func authenticate(
	baseURL url.URL,
	clientID, clientSecret string,
) (string, error) {
	url := baseURL.String() + "/auth"

	authenticateRequest := AuthenticateRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}

	authenticateRequestBytes, err := json.Marshal(authenticateRequest)
	if err != nil {
		return "", fmt.Errorf("error marshalling authenticate request: %w", err)
	}

	payload := bytes.NewReader(authenticateRequestBytes)

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
