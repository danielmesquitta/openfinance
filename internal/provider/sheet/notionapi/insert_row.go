package notionapi

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/danielmesquitta/openfinance-to-sheets/internal/provider/sheet"
)

type insertRowReq struct {
	Parent     insertRowReqParent              `json:"parent"`
	Properties map[string]insertRowReqProperty `json:"properties"`
}

type insertRowReqParent struct {
	DatabaseID string `json:"database_id"`
}

type insertRowReqProperty struct {
	Title    []insertRowReqRichText `json:"title,omitempty"`
	RichText []insertRowReqRichText `json:"rich_text,omitempty"`
	Number   *float64               `json:"number,omitempty"`
	Select   *insertRowReqSelect    `json:"select,omitempty"`
	Date     *insertRowReqDate      `json:"date,omitempty"`
}

type insertRowReqSelect struct {
	Name string `json:"name"`
}

type insertRowReqDate struct {
	Start string `json:"start"`
}

type insertRowReqRichText struct {
	Text insertRowReqText `json:"text"`
}

type insertRowReqText struct {
	Content string `json:"content"`
}

func (c *Client) InsertRow(
	ctx context.Context,
	connectionID, tableID string,
	row sheet.Row,
) error {
	conn, ok := c.conns[connectionID]
	if !ok {
		return errors.New("connection not found for ingest profile " + connectionID)
	}

	requestData, err := insertRowRequest(tableID, row)
	if err != nil {
		return fmt.Errorf("invalid row: %w", err)
	}

	res, err := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+conn.accessToken).
		SetBody(requestData).
		Post("/v1/pages")
	if err != nil {
		return fmt.Errorf(
			"failed to insert row with request data %+v: %w",
			requestData,
			err,
		)
	}

	body := res.Body()
	if res.IsError() {
		return fmt.Errorf(
			"failed to insert row with request data %+v and response %s",
			requestData,
			body,
		)
	}

	return nil
}

func insertRowRequest(tableID string, row sheet.Row) (insertRowReq, error) {
	requestData := insertRowReq{
		Parent:     insertRowReqParent{DatabaseID: tableID},
		Properties: make(map[string]insertRowReqProperty, len(row)),
	}

	for column, cell := range row {
		property, err := insertRowProperty(cell)
		if err != nil {
			return insertRowReq{}, fmt.Errorf("column %q: %w", column, err)
		}
		requestData.Properties[column] = property
	}

	return requestData, nil
}

func insertRowProperty(cell sheet.Cell) (insertRowReqProperty, error) {
	if cell == nil {
		return insertRowReqProperty{}, errors.New("cell is nil")
	}

	richText := func(content string) []insertRowReqRichText {
		return []insertRowReqRichText{{Text: insertRowReqText{Content: content}}}
	}

	switch value := cell.(type) {
	case sheet.TitleCell:
		return insertRowReqProperty{Title: richText(string(value))}, nil
	case sheet.TextCell:
		return insertRowReqProperty{RichText: richText(string(value))}, nil
	case sheet.NumberCell:
		number := float64(value)

		return insertRowReqProperty{Number: &number}, nil
	case sheet.SelectCell:
		return insertRowReqProperty{
			Select: &insertRowReqSelect{Name: formatSelectOption(string(value))},
		}, nil
	case sheet.DateCell:
		return insertRowReqProperty{
			Date: &insertRowReqDate{Start: time.Time(value).Format(time.RFC3339)},
		}, nil
	default:
		return insertRowReqProperty{}, fmt.Errorf("unsupported cell type %T", cell)
	}
}
