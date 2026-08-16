package notionapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/danielmesquitta/openfinance-to-sheets/internal/domain/entity"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/provider/sheet"
)

type retrieveTableResp struct {
	Properties map[string]retrieveTableRespProperty `json:"properties"`
}

type retrieveTableRespProperty struct {
	Type   string                   `json:"type"`
	Select *retrieveTableRespSelect `json:"select"`
}

type retrieveTableRespSelect struct {
	Options []retrieveTableRespSelectOption `json:"options"`
}

type retrieveTableRespSelectOption struct {
	ID    string       `json:"id"`
	Name  string       `json:"name"`
	Color entity.Color `json:"color"`
}

type updateTableReq struct {
	Properties map[string]updateTableReqProperty `json:"properties"`
}

type updateTableReqProperty struct {
	Select *updateTableReqSelect `json:"select,omitempty"`
}

type updateTableReqSelect struct {
	Options []updateTableReqSelectOption `json:"options"`
}

type updateTableReqSelectOption struct {
	ID    string       `json:"id,omitempty"`
	Name  string       `json:"name,omitempty"`
	Color entity.Color `json:"color,omitempty"`
}

func (c *Client) EnsureTableColumns(
	ctx context.Context,
	connectionID string,
	tableID string,
	columns ...sheet.Column,
) error {
	conn, ok := c.conns[connectionID]
	if !ok {
		return errors.New("connection not found for ingest profile " + connectionID)
	}
	if err := (sheet.CreateTableOptions{Columns: columns}).Validate(); err != nil {
		return fmt.Errorf("invalid columns: %w", err)
	}
	for _, column := range columns {
		if column.Type() != sheet.ColumnTypeSelect {
			return fmt.Errorf("column %q has unsupported ensure type %q", column.Name(), column.Type())
		}
	}

	table, err := c.retrieveTable(ctx, conn, tableID)
	if err != nil {
		return err
	}

	updates := make(map[string]updateTableReqProperty)
	for _, column := range columns {
		property, changed, err := ensureSelectProperty(table.Properties, column)
		if err != nil {
			return err
		}
		if changed {
			updates[column.Name()] = property
		}
	}
	if len(updates) == 0 {
		return nil
	}

	requestData := updateTableReq{Properties: updates}
	res, err := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+conn.accessToken).
		SetBody(requestData).
		Patch(fmt.Sprintf("/v1/databases/%s", tableID))
	if err != nil {
		return fmt.Errorf("failed to update database schema: %w", err)
	}
	if res.IsError() {
		return fmt.Errorf("failed to update database schema with response: %s", res.Body())
	}

	return nil
}

func (c *Client) retrieveTable(
	ctx context.Context,
	conn conn,
	tableID string,
) (retrieveTableResp, error) {
	res, err := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+conn.accessToken).
		Get(fmt.Sprintf("/v1/databases/%s", tableID))
	if err != nil {
		return retrieveTableResp{}, fmt.Errorf("failed to retrieve database schema: %w", err)
	}
	if res.IsError() {
		return retrieveTableResp{}, fmt.Errorf("failed to retrieve database schema with response: %s", res.Body())
	}

	var table retrieveTableResp
	if err := json.Unmarshal(res.Body(), &table); err != nil {
		return retrieveTableResp{}, fmt.Errorf("failed to unmarshal database schema: %w", err)
	}

	return table, nil
}

func ensureSelectProperty(
	properties map[string]retrieveTableRespProperty,
	column sheet.Column,
) (updateTableReqProperty, bool, error) {
	existing, exists := properties[column.Name()]
	if !exists {
		return updateTableReqProperty{Select: &updateTableReqSelect{
			Options: newSelectOptions(column.SelectOptions()),
		}}, true, nil
	}
	if existing.Type != string(sheet.ColumnTypeSelect) {
		return updateTableReqProperty{}, false, fmt.Errorf(
			"column %q has type %q, want %q",
			column.Name(),
			existing.Type,
			sheet.ColumnTypeSelect,
		)
	}

	existingOptions := make([]retrieveTableRespSelectOption, 0)
	if existing.Select != nil {
		existingOptions = existing.Select.Options
	}
	options := make([]updateTableReqSelectOption, 0, len(existingOptions)+len(column.SelectOptions()))
	existingNames := make(map[string]struct{}, len(existingOptions))
	for _, option := range existingOptions {
		existingNames[option.Name] = struct{}{}
		preserved := updateTableReqSelectOption{ID: option.ID}
		if option.ID == "" {
			preserved.Name = option.Name
		}
		options = append(options, preserved)
	}

	changed := false
	for _, option := range column.SelectOptions() {
		name := formatSelectOption(option.Name())
		if _, exists := existingNames[name]; exists {
			continue
		}

		existingNames[name] = struct{}{}
		options = append(options, updateTableReqSelectOption{
			Name:  name,
			Color: option.Color(),
		})
		changed = true
	}
	if !changed {
		return updateTableReqProperty{}, false, nil
	}

	return updateTableReqProperty{Select: &updateTableReqSelect{Options: options}}, true, nil
}

func newSelectOptions(options []sheet.SelectOption) []updateTableReqSelectOption {
	result := make([]updateTableReqSelectOption, 0, len(options))
	for _, option := range options {
		result = append(result, updateTableReqSelectOption{
			Name:  formatSelectOption(option.Name()),
			Color: option.Color(),
		})
	}

	return result
}

var _ sheet.Provider = (*Client)(nil)
