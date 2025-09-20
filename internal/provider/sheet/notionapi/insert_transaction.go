package notionapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/pkg/ptr"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
)

type insertRowReq struct {
	Parent     insertRowReqParent     `json:"parent"`
	Properties insertRowReqProperties `json:"properties"`
}

type insertRowReqParent struct {
	DatabaseID string `json:"database_id"`
}

type insertRowReqProperties struct {
	Name           insertRowReqName             `json:"Name"`
	Category       insertRowReqSelector         `json:"Category"`
	Amount         insertRowReqNumber           `json:"Amount"`
	PaymentMethod  insertRowReqSelector         `json:"Payment Method"`
	CardLastDigits insertRowReqRichTextProperty `json:"Card Last Digits"`
	Date           insertRowReqDate             `json:"Date"`
}

type insertRowReqNumber struct {
	Number float64 `json:"number"`
}

type insertRowReqSelector struct {
	Select insertRowReqSelect `json:"select"`
}

type insertRowReqSelect struct {
	Name string `json:"name"`
}

type insertRowReqDate struct {
	Date insertRowReqSubDate `json:"date"`
}

type insertRowReqSubDate struct {
	Start string `json:"start"`
}

type insertRowReqRichText struct {
	Text insertRowReqText `json:"text"`
}

type insertRowReqText struct {
	Content string `json:"content"`
}

type insertRowReqName struct {
	Title []insertRowReqRichText `json:"title"`
}

type insertRowReqRichTextProperty struct {
	RichText []insertRowReqRichText `json:"rich_text"`
}

func (c *Client) InsertTransaction(
	ctx context.Context,
	userID, tableID string,
	transaction entity.Transaction,
) (*sheet.Table, error) {
	conn, ok := c.conns[userID]
	if !ok {
		return nil, errors.New("connection not found for user " + userID)
	}

	requestData := insertRowReq{
		Parent: insertRowReqParent{
			DatabaseID: tableID,
		},
		Properties: insertRowReqProperties{
			Name: insertRowReqName{
				Title: []insertRowReqRichText{
					{
						Text: insertRowReqText{
							Content: transaction.Name,
						},
					},
				},
			},
			Category: insertRowReqSelector{
				Select: insertRowReqSelect{
					Name: string(entity.CategoryUnknown),
				},
			},
			Amount: insertRowReqNumber{
				Number: transaction.Amount,
			},
			PaymentMethod: insertRowReqSelector{
				Select: insertRowReqSelect{
					Name: string(transaction.PaymentMethod),
				},
			},
			CardLastDigits: insertRowReqRichTextProperty{
				RichText: []insertRowReqRichText{
					{
						Text: insertRowReqText{
							Content: ptr.Deref(transaction.CardLastDigits),
						},
					},
				},
			},
			Date: insertRowReqDate{
				Date: insertRowReqSubDate{
					Start: transaction.Date.Format(time.RFC3339),
				},
			},
		},
	}

	if transaction.Category != "" && transaction.Category != entity.CategoryUnknown {
		requestData.Properties.Category = insertRowReqSelector{
			Select: insertRowReqSelect{
				Name: formatSelectOption(string(transaction.Category)),
			},
		}
	}

	res, err := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+conn.accessToken).
		SetBody(requestData).
		Post("/v1/pages")
	if err != nil {
		return nil, fmt.Errorf(
			"failed to insert transaction with request data %+v: %w",
			requestData,
			err,
		)
	}

	body := res.Body()
	if res.IsError() {
		return nil, fmt.Errorf(
			"failed to insert transaction with request data %+v and response %s",
			requestData,
			body,
		)
	}

	data := &sheet.Table{}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal while inserting transaction: %w", err)
	}

	return data, nil
}
