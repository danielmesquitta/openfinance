package notionapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
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

type _conn struct {
	accessToken string
	pageID      string
}

type Client struct {
	baseURL url.URL
	conns   map[string]_conn
}

func NewClient(env *config.Env) *Client {
	baseURL := url.URL{
		Scheme: "https",
		Host:   "api.notion.com",
	}

	conns := map[string]_conn{}
	for _, user := range env.Users {
		conns[user.ID] = _conn{
			accessToken: user.NotionToken,
			pageID:      user.NotionPageID,
		}
	}

	return &Client{
		baseURL: baseURL,
		conns:   conns,
	}
}

func (c *Client) doRequest(
	method, path, token string,
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
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Notion-Version", "2022-06-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res == nil {
		return entity.NewErr("response is nil")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return parseResError(res)
	}

	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&responseData); err != nil {
		return entity.NewErr(err)
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
		return entity.NewErr(fmt.Sprintf(
			"error requesting %s: %s",
			res.Request.URL,
			res.Status,
		))
	}

	return entity.NewErr(fmt.Sprintf(
		"error requesting %s: %s %v",
		res.Request.URL,
		res.Status,
		jsonData,
	))
}

func formatSelectOption(option string) string {
	return strings.ReplaceAll(option, ",", "")
}

var _ sheet.Provider = (*Client)(nil)
