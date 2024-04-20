package meupluggyapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type MeuPluggyAPIClient struct {
	BaseURL      url.URL
	ClientID     string
	ClientSecret string
	Token        string
}

func NewClient(clientID, clientSecret string) *MeuPluggyAPIClient {
	baseURL := url.URL{
		Scheme: "https",
		Host:   "api.pluggy.ai",
	}

	return &MeuPluggyAPIClient{
		BaseURL:      baseURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}
}

type ErrorMessage struct {
	Code            int    `json:"code"`
	Message         string `json:"message"`
	CodeDescription string `json:"codeDescription"`
	Data            any    `json:"data"`
}

func parseResError(res *http.Response) error {
	jsonData := ErrorMessage{}
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&jsonData); err != nil {
		return fmt.Errorf(
			"error requesting %s: %s",
			res.Request.URL,
			res.Status,
		)
	}

	return fmt.Errorf(
		"error requesting %s: %s %v",
		res.Request.URL,
		res.Status,
		jsonData,
	)
}
