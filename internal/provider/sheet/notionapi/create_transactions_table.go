package notionapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
	Name          notionNewTableReqName     `json:"Name"`
	Category      notionNewTableReqCategory `json:"Category"`
	Amount        notionNewTableReqAmount   `json:"Amount"`
	PaymentMethod notionNewTableReqCategory `json:"Payment Method"`
	Date          notionNewTableReqDate     `json:"Date"`
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
	Name  string `json:"name"`
	Color Color  `json:"color"`
}

type notionNewTableReqDate struct {
	Date struct{} `json:"date"`
}

type notionNewTableReqName struct {
	Title struct{} `json:"title"`
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
	dto sheet.CreateTransactionsTableDTO,
) (*sheet.Table, error) {
	conn, ok := c.conns[userID]
	if !ok {
		return nil, errors.New("connection not found for user " + userID)
	}

	categoryOptions := make(
		[]notionNewTableReqSelectOption,
		len(dto.Categories)+1,
	)
	for i, category := range dto.Categories {
		categoryOptions[i] = notionNewTableReqSelectOption{
			Name:  formatSelectOption(string(category)),
			Color: colors[i%len(colors)],
		}
	}
	categoryOptions[len(dto.Categories)] = notionNewTableReqSelectOption{
		Name:  string(sheet.CategoryUnknown),
		Color: Gray,
	}

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
					Content: dto.Title,
				},
			},
		},
		Properties: notionNewTableReqProperties{
			Name: notionNewTableReqName{
				Title: struct{}{},
			},
			Category: notionNewTableReqCategory{
				Select: notionNewTableReqSelect{
					Options: categoryOptions,
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
						{Name: "BOLETO", Color: Yellow},
						{Name: "PIX", Color: Blue},
						{Name: "TED", Color: Green},
						{Name: "CREDIT CARD", Color: Purple},
					},
				},
			},
			Date: notionNewTableReqDate{
				Date: struct{}{},
			},
		},
	}

	res, err := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+conn.accessToken).
		SetBody(requestData).
		Post("/v1/databases")

	if err != nil {
		return nil, fmt.Errorf("failed to create transactions table: %w", err)
	}

	body := res.Body()
	if res.IsError() {
		return nil, fmt.Errorf("error response while creating transactions table: %+v", body)
	}

	data := &sheet.Table{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal while creating transactions table: %w", err)
	}

	return data, nil
}
