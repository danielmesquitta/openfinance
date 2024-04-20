package notionapi

import (
	"net/http"
	"net/url"
)

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

func parseResError(res *http.Response) error {
	return nil
}
