package notionapi

import (
	"strings"

	"github.com/danielmesquitta/openfinance/internal/config"
	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
	"github.com/go-resty/resty/v2"
)

type conn struct {
	accessToken      string
	pageID           string
	colorsByCategory map[entity.Category]entity.Color
}

type Client struct {
	client *resty.Client
	conns  map[string]conn
}

func NewClient(env *config.Env) *Client {
	client := resty.New().
		SetBaseURL("https://api.notion.com").
		SetHeader("Notion-Version", "2022-06-28")

	colorsByCategoryBySyncProfileID := make(
		map[string]map[entity.Category]entity.Color,
		len(env.SyncSettings.SyncProfiles),
	)
	for _, settings := range env.SyncSettings.SyncProfiles {
		colorsByCategoryBySyncProfileID[settings.ID] = settings.ColorsByCategory
	}

	conns := map[string]conn{}
	for _, syncProfile := range env.SyncProfiles {
		conns[syncProfile.ID] = conn{
			accessToken:      syncProfile.NotionToken,
			pageID:           syncProfile.NotionPageID,
			colorsByCategory: colorsByCategoryBySyncProfileID[syncProfile.ID],
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
