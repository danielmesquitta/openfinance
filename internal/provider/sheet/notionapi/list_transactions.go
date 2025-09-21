package notionapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

const maxPageSize = 100

type listTransactionsReq struct {
	StartCursor string                    `json:"start_cursor,omitempty"`
	PageSize    int                       `json:"page_size,omitempty"`
	Sorts       []listTransactionsReqSort `json:"sorts,omitempty"`
}

type listTransactionsReqSort struct {
	Property  string `json:"property,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
	Direction string `json:"direction"`
}

type listTransactionsResp struct {
	Object     string                     `json:"object"`
	Results    []listTransactionsRespPage `json:"results"`
	NextCursor *string                    `json:"next_cursor"`
	HasMore    bool                       `json:"has_more"`
}

type listTransactionsRespPage struct {
	Object         string                             `json:"object"`
	ID             string                             `json:"id"`
	CreatedTime    string                             `json:"created_time"`
	LastEditedTime string                             `json:"last_edited_time"`
	Properties     listTransactionsRespPageProperties `json:"properties"`
}

type listTransactionsRespPageProperties struct {
	Category       listTransactionsRespSelectProperty   `json:"Category"`
	CardLastDigits listTransactionsRespRichTextProperty `json:"Card Last Digits"`
	Amount         listTransactionsRespNumberProperty   `json:"Amount"`
	PaymentMethod  listTransactionsRespSelectProperty   `json:"Payment Method"`
	Date           listTransactionsRespDateProperty     `json:"Date"`
	Name           listTransactionsRespTitleProperty    `json:"Name"`
}

type listTransactionsRespSelectProperty struct {
	ID     string                            `json:"id"`
	Type   string                            `json:"type"`
	Select *listTransactionsRespSelectOption `json:"select"`
}

type listTransactionsRespSelectOption struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type listTransactionsRespRichTextProperty struct {
	ID       string                             `json:"id"`
	Type     string                             `json:"type"`
	RichText []listTransactionsRespRichTextItem `json:"rich_text"`
}

type listTransactionsRespRichTextItem struct {
	Type      string                          `json:"type"`
	Text      listTransactionsRespTextContent `json:"text"`
	PlainText string                          `json:"plain_text"`
}

type listTransactionsRespTextContent struct {
	Content string  `json:"content"`
	Link    *string `json:"link"`
}

type listTransactionsRespNumberProperty struct {
	ID     string   `json:"id"`
	Type   string   `json:"type"`
	Number *float64 `json:"number"`
}

type listTransactionsRespDateProperty struct {
	ID   string                         `json:"id"`
	Type string                         `json:"type"`
	Date *listTransactionsRespDateValue `json:"date"`
}

type listTransactionsRespDateValue struct {
	Start    string  `json:"start"`
	End      *string `json:"end"`
	TimeZone *string `json:"time_zone"`
}

type listTransactionsRespTitleProperty struct {
	ID    string                             `json:"id"`
	Type  string                             `json:"type"`
	Title []listTransactionsRespRichTextItem `json:"title"`
}

func (c *Client) ListTransactions(
	ctx context.Context,
	userID, databaseID string,
) ([]entity.Transaction, error) {
	conn, ok := c.conns[userID]
	if !ok {
		return nil, errors.New("connection not found for user " + userID)
	}

	var allTransactions []entity.Transaction
	var cursor string
	hasMore := true

	for hasMore {
		resp, err := c.queryTransactionsPage(ctx, conn, databaseID, cursor)
		if err != nil {
			return nil, err
		}

		allTransactions = append(allTransactions, c.processTransactionPages(resp.Results)...)

		hasMore = resp.HasMore
		if resp.NextCursor != nil {
			cursor = *resp.NextCursor
		}
	}

	return allTransactions, nil
}

func (c *Client) queryTransactionsPage(
	ctx context.Context,
	conn conn,
	databaseID, cursor string,
) (*listTransactionsResp, error) {
	requestData := listTransactionsReq{
		PageSize: maxPageSize,
		Sorts: []listTransactionsReqSort{
			{
				Timestamp: "created_time",
				Direction: "descending",
			},
		},
	}
	if cursor != "" {
		requestData.StartCursor = cursor
	}

	res, err := c.client.R().
		SetContext(ctx).
		SetHeader("Authorization", "Bearer "+conn.accessToken).
		SetBody(requestData).
		Post(fmt.Sprintf("/v1/databases/%s/query", databaseID))
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %w", err)
	}

	if res.IsError() {
		return nil, fmt.Errorf(
			"failed to query database with response: %s",
			res.Body(),
		)
	}

	var resp listTransactionsResp
	if err := json.Unmarshal(res.Body(), &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp, nil
}

func (c *Client) processTransactionPages(pages []listTransactionsRespPage) []entity.Transaction {
	transactions := make([]entity.Transaction, 0, len(pages))
	for _, page := range pages {
		transaction, err := c.mapPageToTransaction(page)
		if err != nil {
			continue
		}
		transactions = append(transactions, transaction)
	}

	return transactions
}

func (c *Client) mapPageToTransaction(page listTransactionsRespPage) (entity.Transaction, error) {
	var transaction entity.Transaction

	if len(page.Properties.Name.Title) > 0 {
		transaction.Name = page.Properties.Name.Title[0].PlainText
	}

	if page.Properties.Category.Select != nil && page.Properties.Category.Select.Name != "" {
		transaction.Category = entity.Category(page.Properties.Category.Select.Name)
	} else {
		transaction.Category = entity.CategoryUnknown
	}

	if page.Properties.Amount.Number != nil {
		transaction.Amount = *page.Properties.Amount.Number
	}

	if page.Properties.PaymentMethod.Select != nil {
		transaction.PaymentMethod = entity.PaymentMethod(page.Properties.PaymentMethod.Select.Name)
	}

	if len(page.Properties.CardLastDigits.RichText) > 0 &&
		page.Properties.CardLastDigits.RichText[0].PlainText != "" {
		cardLastDigits := page.Properties.CardLastDigits.RichText[0].PlainText
		transaction.CardLastDigits = &cardLastDigits
	}

	if page.Properties.Date.Date != nil {
		parsedTime, err := time.Parse(time.RFC3339, page.Properties.Date.Date.Start)
		if err != nil {
			return transaction, fmt.Errorf(
				"failed to parse date %s: %w",
				page.Properties.Date.Date.Start,
				err,
			)
		}

		transaction.Date = parsedTime
	}

	return transaction, nil
}
