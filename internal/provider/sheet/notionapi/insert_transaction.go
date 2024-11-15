package notionapi

import (
	"encoding/json"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/domain/errs"
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
	Name          insertRowReqName     `json:"Name"`
	Category      insertRowReqSelector `json:"Category"`
	Amount        insertRowReqNumber   `json:"Amount"`
	PaymentMethod insertRowReqSelector `json:"Payment Method"`
	Date          insertRowReqDate     `json:"Date"`
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

func (c *Client) InsertTransaction(
	userID, tableID string,
	transaction entity.Transaction,
) (*sheet.Table, error) {
	conn, ok := c.conns[userID]
	if !ok {
		return nil, errs.New("connection not found for user " + userID)
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
					Name: "Others",
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
			Date: insertRowReqDate{
				Date: insertRowReqSubDate{
					Start: transaction.Date.Format(time.RFC3339),
				},
			},
		},
	}

	if transaction.Category != "" {
		requestData.Properties.Category = insertRowReqSelector{
			Select: insertRowReqSelect{
				Name: formatSelectOption(transaction.Category),
			},
		}
	}

	res, err := c.client.R().
		SetHeader("Authorization", "Bearer "+conn.accessToken).
		SetBody(requestData).
		Post("/v1/pages")
	if err != nil {
		return nil, errs.New(err)
	}

	body := res.Body()
	if statusCode := res.StatusCode(); statusCode < 200 || statusCode >= 300 {
		return nil, errs.New(body)
	}

	data := &sheet.Table{}
	if err := json.Unmarshal(res.Body(), &data); err != nil {
		return nil, errs.New(err)
	}

	return data, nil
}
