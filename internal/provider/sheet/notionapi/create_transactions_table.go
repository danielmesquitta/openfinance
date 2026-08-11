package notionapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
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
	ingestProfileID string,
	title string,
) (entity.Table, error) {
	conn, ok := c.conns[ingestProfileID]
	if !ok {
		return entity.Table{}, errors.New("connection not found for ingest profile " + ingestProfileID)
	}

	requestData := c.getRequestData(conn, title)

	res, err := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+conn.accessToken).
		SetBody(requestData).
		Post("/v1/databases")

	if err != nil {
		return entity.Table{}, fmt.Errorf(
			"failed to create transactions table with request data %+v: %w",
			requestData,
			err,
		)
	}

	body := res.Body()
	if res.IsError() {
		return entity.Table{}, fmt.Errorf(
			"request creating transactions table %+v failed with response %s",
			requestData,
			body,
		)
	}

	data := &createTransactionTableResp{}
	if err := json.Unmarshal(body, &data); err != nil {
		return entity.Table{}, fmt.Errorf("failed to unmarshal while creating transactions table: %w", err)
	}

	if len(data.Title) == 0 {
		return entity.Table{}, errors.New("title is empty")
	}

	table := entity.Table{
		ID:    data.ID,
		Title: data.Title[0].PlainText,
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
					Options: getCategoryOptions(conn.colorsByCategory),
				},
			},
			Amount: createTransactionTableReqAmount{
				Number: createTransactionTableReqNumber{
					Format: "real",
				},
			},
			PaymentMethod: createTransactionTableReqCategory{
				Select: createTransactionTableReqSelect{
					Options: paymentMethodOptions(),
				},
			},
			CardLastDigits: createTransactionTableReqRichText{},
			Date:           createTransactionTableReqDate{},
		},
	}

	return requestData
}

func paymentMethodOptions() []createTransactionTableReqSelectOption {
	options := make([]createTransactionTableReqSelectOption, 0, len(entity.PaymentMethods))
	for _, paymentMethod := range entity.PaymentMethods {
		options = append(options, createTransactionTableReqSelectOption{
			Name:  string(paymentMethod),
			Color: entity.PaymentMethodColors[paymentMethod],
		})
	}

	return options
}

func getCategoryOptions(
	colorsByCategory map[entity.Category]entity.Color,
) []createTransactionTableReqSelectOption {
	categoryOptions := make(
		[]createTransactionTableReqSelectOption,
		0,
		len(colorsByCategory),
	)

	for category, color := range colorsByCategory {
		categoryName := formatSelectOption(string(category))
		categoryOptions = append(categoryOptions, createTransactionTableReqSelectOption{
			Name:  categoryName,
			Color: color,
		})
	}

	return categoryOptions
}
