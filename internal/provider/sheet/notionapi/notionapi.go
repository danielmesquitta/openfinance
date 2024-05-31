package notionapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/danielmesquitta/openfinance/internal/config"
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

func NewClient(env *config.Env) *Client {
	baseURL := url.URL{
		Scheme: "https",
		Host:   "api.notion.com",
	}

	return &Client{
		BaseURL: baseURL,
		Token:   env.NotionToken,
	}
}

func (c *Client) doRequest(
	method, path string,
	requestData any,
	responseData any,
) error {
	jsonRequestData, err := json.Marshal(requestData)
	if err != nil {
		return err
	}
	body := bytes.NewReader(jsonRequestData)

	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return err
	}

	req.Header.Set("accept", "application/json")
	req.Header.Set("content-type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Notion-Version", "2022-06-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return parseResError(res)
	}

	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&responseData); err != nil {
		return err
	}

	return nil
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

func formatSelectOption(option string) string {
	return strings.ReplaceAll(option, ",", "")
}
