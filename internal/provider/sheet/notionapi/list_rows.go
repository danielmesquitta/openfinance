package notionapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
)

const maxPageSize = 100

type listRowsReq struct {
	StartCursor string            `json:"start_cursor,omitempty"`
	PageSize    int               `json:"page_size,omitempty"`
	Sorts       []listRowsReqSort `json:"sorts,omitempty"`
}

type listRowsReqSort struct {
	Property  string `json:"property,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Direction string `json:"direction"`
}

type listRowsResp struct {
	Results    []listRowsRespPage `json:"results"`
	NextCursor *string            `json:"next_cursor"`
	HasMore    bool               `json:"has_more"`
}

type listRowsRespPage struct {
	Properties map[string]listRowsRespProperty `json:"properties"`
}

type listRowsRespProperty struct {
	Type     string                 `json:"type"`
	Title    []listRowsRespRichText `json:"title"`
	RichText []listRowsRespRichText `json:"rich_text"`
	Number   *float64               `json:"number"`
	Select   *listRowsRespSelect    `json:"select"`
	Date     *listRowsRespDate      `json:"date"`
}

type listRowsRespSelect struct {
	Name string `json:"name"`
}

type listRowsRespRichText struct {
	PlainText string `json:"plain_text"`
}

type listRowsRespDate struct {
	Start string `json:"start"`
}

func (c *Client) ListRows(
	ctx context.Context,
	connectionID, tableID string,
) ([]sheet.Row, error) {
	conn, ok := c.conns[connectionID]
	if !ok {
		return nil, errors.New("connection not found for ingest profile " + connectionID)
	}

	var rows []sheet.Row
	var cursor string
	hasMore := true

	for hasMore {
		resp, err := c.queryRowsPage(ctx, conn, tableID, cursor)
		if err != nil {
			return nil, err
		}

		rows = append(rows, processRows(resp.Results)...)

		hasMore = resp.HasMore
		if resp.NextCursor != nil {
			cursor = *resp.NextCursor
		}
	}

	return rows, nil
}

func (c *Client) queryRowsPage(
	ctx context.Context,
	conn conn,
	tableID, cursor string,
) (*listRowsResp, error) {
	requestData := listRowsReq{
		PageSize: maxPageSize,
		Sorts: []listRowsReqSort{{
			Timestamp: "created_time",
			Direction: "descending",
		}},
	}
	if cursor != "" {
		requestData.StartCursor = cursor
	}

	res, err := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+conn.accessToken).
		SetBody(requestData).
		Post(fmt.Sprintf("/v1/databases/%s/query", tableID))
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}
	if res.IsError() {
		return nil, fmt.Errorf("failed to query database with response: %s", res.Body())
	}

	var resp listRowsResp
	if err := json.Unmarshal(res.Body(), &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

func processRows(pages []listRowsRespPage) []sheet.Row {
	rows := make([]sheet.Row, 0, len(pages))
	for _, page := range pages {
		row, err := mapPageToRow(page)
		if err != nil {
			continue
		}
		rows = append(rows, row)
	}

	return rows
}

func mapPageToRow(page listRowsRespPage) (sheet.Row, error) {
	row := make(sheet.Row, len(page.Properties))
	for name, property := range page.Properties {
		cell, ok, err := mapPropertyToCell(property)
		if err != nil {
			return nil, err
		}
		if ok {
			row[name] = cell
		}
	}

	return row, nil
}

func mapPropertyToCell(property listRowsRespProperty) (sheet.Cell, bool, error) {
	switch property.Type {
	case "title":
		return sheet.TitleCell(firstPlainText(property.Title)), true, nil
	case "rich_text":
		return sheet.TextCell(firstPlainText(property.RichText)), true, nil
	case "number":
		if property.Number == nil {
			return nil, false, nil
		}

		return sheet.NumberCell(*property.Number), true, nil
	case "select":
		if property.Select == nil {
			return nil, false, nil
		}

		return sheet.SelectCell(property.Select.Name), true, nil
	case "date":
		if property.Date == nil {
			return nil, false, nil
		}
		value, err := time.Parse(time.RFC3339, property.Date.Start)
		if err != nil {
			return nil, false, fmt.Errorf("failed to parse date %s: %w", property.Date.Start, err)
		}

		return sheet.DateCell(value), true, nil
	default:
		return nil, false, nil
	}
}

func firstPlainText(items []listRowsRespRichText) string {
	if len(items) == 0 {
		return ""
	}

	return items[0].PlainText
}
