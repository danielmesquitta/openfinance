package notionapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/danielmesquitta/openfinance-to-sheets/internal/provider/sheet"
)

type createTableReq struct {
	Parent     createTableReqParent              `json:"parent"`
	Icon       *createTableReqIcon               `json:"icon,omitempty"`
	Title      []createTableReqTitle             `json:"title"`
	Properties map[string]createTableReqProperty `json:"properties"`
}

type createTableReqIcon struct {
	Type  string `json:"type"`
	Emoji string `json:"emoji"`
}

type createTableReqParent struct {
	Type   string `json:"type"`
	PageID string `json:"page_id"`
}

type createTableReqProperty struct {
	Title    *struct{}             `json:"title,omitempty"`
	RichText *struct{}             `json:"rich_text,omitempty"`
	Number   *createTableReqNumber `json:"number,omitempty"`
	Select   *createTableReqSelect `json:"select,omitempty"`
	Date     *struct{}             `json:"date,omitempty"`
}

type createTableReqNumber struct {
	Format string `json:"format,omitempty"`
}

type createTableReqSelect struct {
	Options []createTableReqSelectOption `json:"options"`
}

type createTableReqSelectOption struct {
	Name  string `json:"name"`
	Color string `json:"color,omitempty"`
}

type createTableReqTitle struct {
	Type string             `json:"type"`
	Text createTableReqText `json:"text"`
}

type createTableReqText struct {
	Content string `json:"content"`
}

type createTableResp struct {
	ID    string                 `json:"id"`
	Title []createTableRespTitle `json:"title"`
}

type createTableRespTitle struct {
	PlainText string `json:"plain_text"`
}

func (c *Client) CreateTable(
	ctx context.Context,
	connectionID string,
	title string,
	options ...sheet.CreateTableOption,
) (sheet.Table, error) {
	conn, ok := c.conns[connectionID]
	if !ok {
		return sheet.Table{}, errors.New("connection not found for ingest profile " + connectionID)
	}

	createOptions := sheet.CreateTableOptions{}
	for _, option := range options {
		if option != nil {
			option(&createOptions)
		}
	}
	if err := createOptions.Validate(); err != nil {
		return sheet.Table{}, fmt.Errorf("invalid table options: %w", err)
	}

	requestData, err := createTableRequest(conn.pageID, title, createOptions)
	if err != nil {
		return sheet.Table{}, fmt.Errorf("invalid table options: %w", err)
	}

	res, err := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+conn.accessToken).
		SetBody(requestData).
		Post("/v1/databases")
	if err != nil {
		return sheet.Table{}, fmt.Errorf(
			"failed to create table with request data %+v: %w",
			requestData,
			err,
		)
	}

	body := res.Body()
	if res.IsError() {
		return sheet.Table{}, fmt.Errorf(
			"request creating table %+v failed with response %s",
			requestData,
			body,
		)
	}

	data := createTableResp{}
	if err := json.Unmarshal(body, &data); err != nil {
		return sheet.Table{}, fmt.Errorf("failed to unmarshal while creating table: %w", err)
	}
	if len(data.Title) == 0 {
		return sheet.Table{}, errors.New("title is empty")
	}

	return sheet.Table{ID: data.ID, Title: data.Title[0].PlainText}, nil
}

func createTableRequest(
	pageID, title string,
	options sheet.CreateTableOptions,
) (createTableReq, error) {
	requestData := createTableReq{
		Parent: createTableReqParent{Type: "page_id", PageID: pageID},
		Title: []createTableReqTitle{{
			Type: "text",
			Text: createTableReqText{Content: title},
		}},
		Properties: make(map[string]createTableReqProperty, len(options.Columns)),
	}
	if options.Icon != "" {
		requestData.Icon = &createTableReqIcon{Type: "emoji", Emoji: options.Icon}
	}

	for _, column := range options.Columns {
		property, err := createTableProperty(column)
		if err != nil {
			return createTableReq{}, fmt.Errorf("column %q: %w", column.Name(), err)
		}
		requestData.Properties[column.Name()] = property
	}

	return requestData, nil
}

func createTableProperty(column sheet.Column) (createTableReqProperty, error) {
	empty := &struct{}{}
	switch column.Type() {
	case sheet.ColumnTypeTitle:
		return createTableReqProperty{Title: empty}, nil
	case sheet.ColumnTypeText:
		return createTableReqProperty{RichText: empty}, nil
	case sheet.ColumnTypeNumber:
		number := &createTableReqNumber{}
		switch column.Currency() {
		case "":
		case "BRL":
			number.Format = "real"
		default:
			return createTableReqProperty{}, fmt.Errorf("unsupported currency %q", column.Currency())
		}

		return createTableReqProperty{Number: number}, nil
	case sheet.ColumnTypeSelect:
		selectOptions := column.SelectOptions()
		options := make([]createTableReqSelectOption, 0, len(selectOptions))
		for _, option := range selectOptions {
			options = append(options, createTableReqSelectOption{
				Name:  formatSelectOption(option.Name()),
				Color: string(option.Color()),
			})
		}

		return createTableReqProperty{Select: &createTableReqSelect{Options: options}}, nil
	case sheet.ColumnTypeDate:
		return createTableReqProperty{Date: empty}, nil
	default:
		return createTableReqProperty{}, fmt.Errorf("unsupported column type %q", column.Type())
	}
}
