package notionapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
)

type notionNewTableReq struct {
	Parent     notionNewTableReqParent     `json:"parent"`
	Icon       notionNewTableReqIcon       `json:"icon"`
	Title      []notionNewTableReqTitle    `json:"title"`
	Properties notionNewTableReqProperties `json:"properties"`
}

type notionNewTableReqIcon struct {
	Type  string `json:"type"`
	Emoji string `json:"emoji"`
}

type notionNewTableReqParent struct {
	Type   string `json:"type"`
	PageID string `json:"page_id"`
}

type notionNewTableReqProperties struct {
	Name           notionNewTableReqName     `json:"Name"`
	Category       notionNewTableReqCategory `json:"Category"`
	Amount         notionNewTableReqAmount   `json:"Amount"`
	PaymentMethod  notionNewTableReqCategory `json:"Payment Method"`
	CardLastDigits notionNewTableReqRichText `json:"Card Last Digits"`
	Date           notionNewTableReqDate     `json:"Date"`
}

type notionNewTableReqAmount struct {
	Number notionNewTableReqNumber `json:"number"`
}

type notionNewTableReqNumber struct {
	Format string `json:"format"`
}

type notionNewTableReqCategory struct {
	Select notionNewTableReqSelect `json:"select"`
}

type notionNewTableReqSelect struct {
	Options []notionNewTableReqSelectOption `json:"options"`
}

type notionNewTableReqSelectOption struct {
	Name  string       `json:"name"`
	Color entity.Color `json:"color"`
}

type notionNewTableReqDate struct {
	Date struct{} `json:"date"`
}

type notionNewTableReqName struct {
	Title struct{} `json:"title"`
}

type notionNewTableReqRichText struct {
	RichText struct{} `json:"rich_text"`
}

type notionNewTableReqTitle struct {
	Type string                `json:"type"`
	Text notionNewTableReqText `json:"text"`
}

type notionNewTableReqText struct {
	Content string `json:"content"`
}

func (c *Client) CreateTransactionsTable(
	ctx context.Context,
	userID string,
	title string,
) (*sheet.Table, error) {
	conn, ok := c.conns[userID]
	if !ok {
		return nil, errors.New("connection not found for user " + userID)
	}

	requestData := c.getRequestData(conn, title)

	res, err := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+conn.accessToken).
		SetBody(requestData).
		Post("/v1/databases")

	if err != nil {
		return nil, fmt.Errorf(
			"failed to create transactions table with request data %+v: %w",
			requestData,
			err,
		)
	}

	body := res.Body()
	if res.IsError() {
		return nil, fmt.Errorf(
			"request creating transactions table %+v failed with response %s",
			requestData,
			body,
		)
	}

	data := &sheet.Table{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal while creating transactions table: %w", err)
	}

	return data, nil
}

func (c *Client) getRequestData(
	conn conn,
	title string,
) notionNewTableReq {
	requestData := notionNewTableReq{
		Parent: notionNewTableReqParent{
			Type:   "page_id",
			PageID: conn.pageID,
		},
		Icon: notionNewTableReqIcon{
			Type:  "emoji",
			Emoji: "💸",
		},
		Title: []notionNewTableReqTitle{
			{
				Type: "text",
				Text: notionNewTableReqText{
					Content: title,
				},
			},
		},
		Properties: notionNewTableReqProperties{
			Name: notionNewTableReqName{},
			Category: notionNewTableReqCategory{
				Select: notionNewTableReqSelect{
					Options: c.getCategoryOptions(),
				},
			},
			Amount: notionNewTableReqAmount{
				Number: notionNewTableReqNumber{
					Format: "real",
				},
			},
			PaymentMethod: notionNewTableReqCategory{
				Select: notionNewTableReqSelect{
					Options: []notionNewTableReqSelectOption{
						{Name: "BOLETO", Color: entity.Yellow},
						{Name: "PIX", Color: entity.Blue},
						{Name: "TED", Color: entity.Green},
						{Name: "CREDIT CARD", Color: entity.Purple},
					},
				},
			},
			CardLastDigits: notionNewTableReqRichText{},
			Date:           notionNewTableReqDate{},
		},
	}

	return requestData
}

func (c *Client) getCategoryOptions() []notionNewTableReqSelectOption {
	categoryOptions := make(
		[]notionNewTableReqSelectOption,
		0,
		len(c.env.ColorsByCategory),
	)

	for category, color := range c.env.ColorsByCategory {
		categoryName := formatSelectOption(string(category))
		categoryOptions = append(categoryOptions, notionNewTableReqSelectOption{
			Name:  categoryName,
			Color: color,
		})
	}

	return categoryOptions
}
