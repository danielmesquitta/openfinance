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

	colorsByCategoryByIngestProfileID := make(
		map[string]map[entity.Category]entity.Color,
		len(env.IngestSettings.IngestProfiles),
	)
	for _, settings := range env.IngestSettings.IngestProfiles {
		colorsByCategoryByIngestProfileID[settings.ID] = settings.ColorsByCategory
	}

	conns := map[string]conn{}
	for _, ingestProfile := range env.IngestProfiles {
		conns[ingestProfile.ID] = conn{
			accessToken:      ingestProfile.NotionToken,
			pageID:           ingestProfile.NotionPageID,
			colorsByCategory: colorsByCategoryByIngestProfileID[ingestProfile.ID],
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
