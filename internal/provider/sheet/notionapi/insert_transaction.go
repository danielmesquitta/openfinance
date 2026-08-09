package notionapi

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

type insertTransactionReq struct {
	Parent     insertTransactionReqParent     `json:"parent"`
	Properties insertTransactionReqProperties `json:"properties"`
}

type insertTransactionReqParent struct {
	DatabaseID string `json:"database_id"`
}

type insertTransactionReqProperties struct {
	Name           insertTransactionReqName             `json:"Name"`
	Category       insertTransactionReqSelector         `json:"Category"`
	Amount         insertTransactionReqNumber           `json:"Amount"`
	PaymentMethod  insertTransactionReqSelector         `json:"Payment Method"`
	CardLastDigits insertTransactionReqRichTextProperty `json:"Card Last Digits"`
	Date           insertTransactionReqDate             `json:"Date"`
}

type insertTransactionReqNumber struct {
	Number float64 `json:"number"`
}

type insertTransactionReqSelector struct {
	Select insertTransactionReqSelect `json:"select"`
}

type insertTransactionReqSelect struct {
	Name string `json:"name"`
}

type insertTransactionReqDate struct {
	Date insertTransactionReqSubDate `json:"date"`
}

type insertTransactionReqSubDate struct {
	Start string `json:"start"`
}

type insertTransactionReqRichText struct {
	Text insertTransactionReqText `json:"text"`
}

type insertTransactionReqText struct {
	Content string `json:"content"`
}

type insertTransactionReqName struct {
	Title []insertTransactionReqRichText `json:"title"`
}

type insertTransactionReqRichTextProperty struct {
	RichText []insertTransactionReqRichText `json:"rich_text"`
}

func (c *Client) InsertTransaction(
	ctx context.Context,
	syncProfileID, tableID string,
	transaction entity.Transaction,
) error {
	conn, ok := c.conns[syncProfileID]
	if !ok {
		return errors.New("connection not found for sync profile " + syncProfileID)
	}

	cardLastDigits := ""
	if transaction.CardLastDigits != nil {
		cardLastDigits = *transaction.CardLastDigits
	}

	requestData := insertTransactionReq{
		Parent: insertTransactionReqParent{
			DatabaseID: tableID,
		},
		Properties: insertTransactionReqProperties{
			Name: insertTransactionReqName{
				Title: []insertTransactionReqRichText{
					{
						Text: insertTransactionReqText{
							Content: transaction.Name,
						},
					},
				},
			},
			Category: insertTransactionReqSelector{
				Select: insertTransactionReqSelect{
					Name: formatSelectOption(string(transaction.Category)),
				},
			},
			Amount: insertTransactionReqNumber{
				Number: transaction.Amount,
			},
			PaymentMethod: insertTransactionReqSelector{
				Select: insertTransactionReqSelect{
					Name: string(transaction.PaymentMethod),
				},
			},
			CardLastDigits: insertTransactionReqRichTextProperty{
				RichText: []insertTransactionReqRichText{
					{
						Text: insertTransactionReqText{
							Content: cardLastDigits,
						},
					},
				},
			},
			Date: insertTransactionReqDate{
				Date: insertTransactionReqSubDate{
					Start: transaction.Date.Format(time.RFC3339),
				},
			},
		},
	}

	res, err := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+conn.accessToken).
		SetBody(requestData).
		Post("/v1/pages")
	if err != nil {
		return fmt.Errorf(
			"failed to insert transaction with request data %+v: %w",
			requestData,
			err,
		)
	}

	body := res.Body()
	if res.IsError() {
		return fmt.Errorf(
			"failed to insert transaction with request data %+v and response %s",
			requestData,
			body,
		)
	}

	return nil
}
