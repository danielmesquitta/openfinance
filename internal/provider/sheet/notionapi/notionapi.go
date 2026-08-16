package notionapi

import (
	"strings"

	"github.com/danielmesquitta/openfinance-to-sheets/internal/config"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/provider/sheet"
	"github.com/go-resty/resty/v2"
)

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
	for _, ingestProfile := range env.IngestProfiles {
		conns[ingestProfile.ID] = conn{
			accessToken: ingestProfile.NotionToken,
			pageID:      ingestProfile.NotionPageID,
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
