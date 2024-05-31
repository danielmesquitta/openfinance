package meupluggyapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Client struct {
	BaseURL url.URL
	Token   string
}

func NewClient(clientID, clientSecret string) *Client {
	baseURL := url.URL{
		Scheme: "https",
		Host:   "api.pluggy.ai",
	}

	token, err := authenticate(baseURL, clientID, clientSecret)
	if err != nil {
		panic(err)
	}

	return &Client{
		BaseURL: baseURL,
		Token:   token,
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
