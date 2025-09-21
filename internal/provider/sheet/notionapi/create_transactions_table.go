package notionapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
)

type createTransactionTableReq struct {
	Parent     createTransactionTableReqParent     `json:"parent"`
	Icon       createTransactionTableReqIcon       `json:"icon"`
	Title      []createTransactionTableReqTitle    `json:"title"`
	Properties createTransactionTableReqProperties `json:"properties"`
}

type createTransactionTableReqIcon struct {
	Type  string `json:"type"`
	Emoji string `json:"emoji"`
}

type createTransactionTableReqParent struct {
	Type   string `json:"type"`
	PageID string `json:"page_id"`
}

type createTransactionTableReqProperties struct {
	Name           createTransactionTableReqName     `json:"Name"`
	Category       createTransactionTableReqCategory `json:"Category"`
	Amount         createTransactionTableReqAmount   `json:"Amount"`
	PaymentMethod  createTransactionTableReqCategory `json:"Payment Method"`
	CardLastDigits createTransactionTableReqRichText `json:"Card Last Digits"`
	Date           createTransactionTableReqDate     `json:"Date"`
}

type createTransactionTableReqAmount struct {
	Number createTransactionTableReqNumber `json:"number"`
}

type createTransactionTableReqNumber struct {
	Format string `json:"format"`
}

type createTransactionTableReqCategory struct {
	Select createTransactionTableReqSelect `json:"select"`
}

type createTransactionTableReqSelect struct {
	Options []createTransactionTableReqSelectOption `json:"options"`
}

type createTransactionTableReqSelectOption struct {
	Name  string       `json:"name"`
	Color entity.Color `json:"color"`
}

type createTransactionTableReqDate struct {
	Date struct{} `json:"date"`
}

type createTransactionTableReqName struct {
	Title struct{} `json:"title"`
}

type createTransactionTableReqRichText struct {
	RichText struct{} `json:"rich_text"`
}

type createTransactionTableReqTitle struct {
	Type string                        `json:"type"`
	Text createTransactionTableReqText `json:"text"`
}

type createTransactionTableReqText struct {
	Content string `json:"content"`
}

type createTransactionTableResp struct {
	ID       string                            `json:"id"`
	Title    []createTransactionTableRespTitle `json:"title"`
	Archived bool                              `json:"archived"`
	InTrash  bool                              `json:"in_trash"`
}

type createTransactionTableRespTitle struct {
	PlainText string `json:"plain_text"`
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

	data := &createTransactionTableResp{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal while creating transactions table: %w", err)
	}

	if len(data.Title) == 0 {
		return nil, errors.New("title is empty")
	}

	table := &sheet.Table{
		ID:       data.ID,
		Title:    &data.Title[0].PlainText,
		Archived: data.Archived,
		InTrash:  data.InTrash,
	}

	return table, nil
}

func (c *Client) getRequestData(
	conn conn,
	title string,
) createTransactionTableReq {
	requestData := createTransactionTableReq{
		Parent: createTransactionTableReqParent{
			Type:   "page_id",
			PageID: conn.pageID,
		},
		Icon: createTransactionTableReqIcon{
			Type:  "emoji",
			Emoji: "💸",
		},
		Title: []createTransactionTableReqTitle{
			{
				Type: "text",
				Text: createTransactionTableReqText{
					Content: title,
				},
			},
		},
		Properties: createTransactionTableReqProperties{
			Name: createTransactionTableReqName{},
			Category: createTransactionTableReqCategory{
				Select: createTransactionTableReqSelect{
					Options: c.getCategoryOptions(),
				},
			},
			Amount: createTransactionTableReqAmount{
				Number: createTransactionTableReqNumber{
					Format: "real",
				},
			},
			PaymentMethod: createTransactionTableReqCategory{
				Select: createTransactionTableReqSelect{
					Options: []createTransactionTableReqSelectOption{
						{Name: "BOLETO", Color: entity.Yellow},
						{Name: "PIX", Color: entity.Blue},
						{Name: "TED", Color: entity.Green},
						{Name: "CREDIT CARD", Color: entity.Purple},
					},
				},
			},
			CardLastDigits: createTransactionTableReqRichText{},
			Date:           createTransactionTableReqDate{},
		},
	}

	return requestData
}

func (c *Client) getCategoryOptions() []createTransactionTableReqSelectOption {
	categoryOptions := make(
		[]createTransactionTableReqSelectOption,
		0,
		len(c.env.ColorsByCategory),
	)

	for category, color := range c.env.ColorsByCategory {
		categoryName := formatSelectOption(string(category))
		categoryOptions = append(categoryOptions, createTransactionTableReqSelectOption{
			Name:  categoryName,
			Color: color,
		})
	}

	return categoryOptions
}
