package notionapi

import (
	"net/http"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
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
	Name          insertRowReqName        `json:"Name"`
	Description   insertRowReqDescription `json:"Description"`
	Category      insertRowReqSelector    `json:"Category"`
	Amount        insertRowReqNumber      `json:"Amount"`
	PaymentMethod insertRowReqSelector    `json:"Payment Method"`
	Date          insertRowReqDate        `json:"Date"`
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

type insertRowReqDescription struct {
	RichText []insertRowReqRichText `json:"rich_text"`
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

func (c *Client) InsertRow(
	databaseID string,
	transaction entity.Transaction,
) (*sheet.Table, error) {
	url := c.BaseURL
	url.Path = "/v1/pages"

	requestData := insertRowReq{
		Parent: insertRowReqParent{
			DatabaseID: databaseID,
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
			Description: insertRowReqDescription{
				RichText: []insertRowReqRichText{
					{
						Text: insertRowReqText{
							Content: transaction.Description,
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

	responseData := &sheet.Table{}
	if err := c.doRequest(
		http.MethodPost,
		url.String(),
		requestData,
		responseData,
	); err != nil {
		return nil, err
	}

	return responseData, nil
}
