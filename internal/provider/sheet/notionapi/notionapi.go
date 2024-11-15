package notionapi

import (
	"strings"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
	"github.com/go-resty/resty/v2"
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

type conn struct {
	accessToken string
	pageID      string
}

type Client struct {
	client *resty.Client
	conns  map[string]conn
}

func NewClient(env *config.Env) *Client {
	client := resty.New().
		SetBaseURL("https://api.notion.com").
		SetHeader("Notion-Version", "2022-06-28")

	conns := map[string]conn{}
	for _, user := range env.Users {
		conns[user.ID] = conn{
			accessToken: user.NotionToken,
			pageID:      user.NotionPageID,
		}
	}

	return &Client{
		client: client,
		conns:  conns,
	}
}

func formatSelectOption(option string) string {
	return strings.ReplaceAll(option, ",", "")
}

var _ sheet.Provider = (*Client)(nil)
