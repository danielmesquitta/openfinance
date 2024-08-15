package notionapi

import (
	"net/http"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
)

type _insertRowReq struct {
	Parent     _insertRowReqParent     `json:"parent"`
	Properties _insertRowReqProperties `json:"properties"`
}

type _insertRowReqParent struct {
	DatabaseID string `json:"database_id"`
}

type _insertRowReqProperties struct {
	Name          _insertRowReqName        `json:"Name"`
	Description   _insertRowReqDescription `json:"Description"`
	Category      _insertRowReqSelector    `json:"Category"`
	Amount        _insertRowReqNumber      `json:"Amount"`
	PaymentMethod _insertRowReqSelector    `json:"Payment Method"`
	Date          _insertRowReqDate        `json:"Date"`
}

type _insertRowReqNumber struct {
	Number float64 `json:"number"`
}

type _insertRowReqSelector struct {
	Select _insertRowReqSelect `json:"select"`
}

type _insertRowReqSelect struct {
	Name string `json:"name"`
}

type _insertRowReqDate struct {
	Date _insertRowReqSubDate `json:"date"`
}

type _insertRowReqSubDate struct {
	Start string `json:"start"`
}

type _insertRowReqDescription struct {
	RichText []_insertRowReqRichText `json:"rich_text"`
}

type _insertRowReqRichText struct {
	Text _insertRowReqText `json:"text"`
}

type _insertRowReqText struct {
	Content string `json:"content"`
}

type _insertRowReqName struct {
	Title []_insertRowReqRichText `json:"title"`
}

func (c *Client) InsertTransaction(
	userID, tableID string,
	transaction entity.Transaction,
) (*sheet.Table, error) {
	conn, ok := c.conns[userID]
	if !ok {
		return nil, entity.NewErr("connection not found for user " + userID)
	}

	url := c.baseURL
	url.Path = "/v1/pages"

	requestData := _insertRowReq{
		Parent: _insertRowReqParent{
			DatabaseID: tableID,
		},
		Properties: _insertRowReqProperties{
			Name: _insertRowReqName{
				Title: []_insertRowReqRichText{
					{
						Text: _insertRowReqText{
							Content: transaction.Name,
						},
					},
				},
			},
			Description: _insertRowReqDescription{
				RichText: []_insertRowReqRichText{
					{
						Text: _insertRowReqText{
							Content: transaction.Description,
						},
					},
				},
			},
			Category: _insertRowReqSelector{
				Select: _insertRowReqSelect{
					Name: "Others",
				},
			},
			Amount: _insertRowReqNumber{
				Number: transaction.Amount,
			},
			PaymentMethod: _insertRowReqSelector{
				Select: _insertRowReqSelect{
					Name: string(transaction.PaymentMethod),
				},
			},
			Date: _insertRowReqDate{
				Date: _insertRowReqSubDate{
					Start: transaction.Date.Format(time.RFC3339),
				},
			},
		},
	}

	if transaction.Category != "" {
		requestData.Properties.Category = _insertRowReqSelector{
			Select: _insertRowReqSelect{
				Name: formatSelectOption(transaction.Category),
			},
		}
	}

	responseData := &sheet.Table{}
	if err := c.doRequest(
		http.MethodPost,
		url.String(),
		conn.accessToken,
		requestData,
		responseData,
	); err != nil {
		return nil, entity.NewErr(err)
	}

	return responseData, nil
}
