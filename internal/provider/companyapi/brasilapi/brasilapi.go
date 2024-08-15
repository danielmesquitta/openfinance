package brasilapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Client struct {
	BaseURL url.URL
}

func NewClient() *Client {
	baseURL := url.URL{
		Scheme: "https",
		Host:   "brasilapi.com.br",
	}

	return &Client{
		BaseURL: baseURL,
	}
}

type ErrorMessage struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Name    string `json:"name"`
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
