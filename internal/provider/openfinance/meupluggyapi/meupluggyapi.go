package meupluggyapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/danielmesquitta/openfinance/internal/config"
)

type Client struct {
	BaseURL      url.URL
	ClientID     string
	ClientSecret string
	Token        string
}

func NewClient(env *config.Env) *Client {
	baseURL := url.URL{
		Scheme: "https",
		Host:   "api.pluggy.ai",
	}

	return &Client{
		BaseURL:      baseURL,
		ClientID:     env.MeuPluggyClientID,
		ClientSecret: env.MeuPluggyClientSecret,
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
