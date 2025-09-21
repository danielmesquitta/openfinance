package notionapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
)

type listTablesResp struct {
	Results    []listTablesResult `json:"results"`
	NextCursor *string            `json:"next_cursor"`
	HasMore    bool               `json:"has_more"`
}

type listTablesResult struct {
	ID             string                   `json:"id"`
	CreatedTime    time.Time                `json:"created_time"`
	LastEditedTime time.Time                `json:"last_edited_time"`
	HasChildren    bool                     `json:"has_children"`
	Archived       bool                     `json:"archived"`
	InTrash        bool                     `json:"in_trash"`
	ChildDatabase  *listTablesChildDatabase `json:"child_database,omitempty"`
}

type listTablesChildDatabase struct {
	Title string `json:"title"`
}

func (c *Client) ListTables(
	ctx context.Context,
	userID string,
) ([]sheet.Table, error) {
	conn, ok := c.conns[userID]
	if !ok {
		return nil, errors.New("connection not found for user " + userID)
	}

	var allTables []sheet.Table
	var cursor string
	hasMore := true

	for hasMore {
		data, err := c.fetchTablesPage(ctx, conn, cursor)
		if err != nil {
			return nil, err
		}

		allTables = append(allTables, c.extractTablesFromResults(data.Results)...)

		hasMore = data.HasMore
		if data.NextCursor != nil {
			cursor = *data.NextCursor
		}
	}

	return allTables, nil
}

func (c *Client) fetchTablesPage(
	ctx context.Context,
	conn conn,
	cursor string,
) (*listTablesResp, error) {
	queryParams := map[string]string{"page_size": "100"}
	if cursor != "" {
		queryParams["start_cursor"] = cursor
	}

	res, err := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+conn.accessToken).
		SetQueryParams(queryParams).
		Get(fmt.Sprintf("/v1/blocks/%s/children", conn.pageID))
	if err != nil {
		return nil, fmt.Errorf("failed to list tables: %w", err)
	}

	if res.IsError() {
		return nil, fmt.Errorf("failed to list tables: %s", res.Body())
	}

	data := listTablesResp{}
	if err := json.Unmarshal(res.Body(), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal while listing tables: %w", err)
	}

	return &data, nil
}

func (c *Client) extractTablesFromResults(results []listTablesResult) []sheet.Table {
	var tables []sheet.Table
	for _, result := range results {
		if result.ChildDatabase != nil {
			table := sheet.Table{
				ID:       result.ID,
				Title:    &result.ChildDatabase.Title,
				Archived: result.Archived,
				InTrash:  result.InTrash,
			}
			tables = append(tables, table)
		}
	}

	return tables
}
