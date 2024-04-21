package notionapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Color string

const (
	Blue      Color = "blue"
	Brown     Color = "brown"
	Gray      Color = "default"
	LightGray Color = "gray"
	Green     Color = "green"
	Orange    Color = "orange"
	Pink      Color = "pink"
	Purple    Color = "purple"
	Red       Color = "red"
	Yellow    Color = "yellow"
)

var colors = []Color{
	Blue,
	Red,
	Green,
	Purple,
	Yellow,
	Pink,
	Orange,
	LightGray,
	Brown,
	Gray,
}

type Client struct {
	BaseURL url.URL
	Token   string
}

func NewClient(token string) *Client {
	baseURL := url.URL{
		Scheme: "https",
		Host:   "api.notion.com",
	}

	return &Client{
		BaseURL: baseURL,
		Token:   token,
	}
}

func setHeaders(req *http.Request, token string) {
	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Notion-Version", "2022-06-28")
}

type ErrorMessage struct {
	Object  string `json:"object"`
	Status  int64  `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
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
