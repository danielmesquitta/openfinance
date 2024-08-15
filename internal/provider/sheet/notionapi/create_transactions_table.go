package notionapi

import (
	"net/http"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
)

type _notionNewTableReq struct {
	Parent     _notionNewTableReqParent     `json:"parent"`
	Icon       _notionNewTableReqIcon       `json:"icon"`
	Title      []_notionNewTableReqTitle    `json:"title"`
	Properties _notionNewTableReqProperties `json:"properties"`
}

type _notionNewTableReqIcon struct {
	Type  string `json:"type"`
	Emoji string `json:"emoji"`
}

type _notionNewTableReqParent struct {
	Type   string `json:"type"`
	PageID string `json:"page_id"`
}

type _notionNewTableReqProperties struct {
	Name          _notionNewTableReqName        `json:"Name"`
	Description   _notionNewTableReqDescription `json:"Description"`
	Category      _notionNewTableReqCategory    `json:"Category"`
	Amount        _notionNewTableReqAmount      `json:"Amount"`
	PaymentMethod _notionNewTableReqCategory    `json:"Payment Method"`
	Date          _notionNewTableReqDate        `json:"Date"`
}

type _notionNewTableReqAmount struct {
	Number _notionNewTableReqNumber `json:"number"`
}

type _notionNewTableReqNumber struct {
	Format string `json:"format"`
}

type _notionNewTableReqCategory struct {
	Select _notionNewTableReqSelect `json:"select"`
}

type _notionNewTableReqSelect struct {
	Options []_notionNewTableReqSelectOption `json:"options"`
}

type _notionNewTableReqSelectOption struct {
	Name  string `json:"name"`
	Color Color  `json:"color"`
}

type _notionNewTableReqDate struct {
	Date struct{} `json:"date"`
}

type _notionNewTableReqDescription struct {
	RichText struct{} `json:"rich_text"`
}

type _notionNewTableReqName struct {
	Title struct{} `json:"title"`
}

type _notionNewTableReqTitle struct {
	Type string                 `json:"type"`
	Text _notionNewTableReqText `json:"text"`
}

type _notionNewTableReqText struct {
	Content string `json:"content"`
}

func (c *Client) CreateTransactionsTable(
	userID string,
	dto sheet.CreateTransactionsTableDTO,
) (*sheet.Table, error) {
	conn, ok := c.conns[userID]
	if !ok {
		return nil, entity.NewErr("connection not found for user " + userID)
	}

	url := c.baseURL
	url.Path = "/v1/databases"

	categoryOptions := make(
		[]_notionNewTableReqSelectOption,
		len(dto.Categories)+1,
	)
	for i, category := range dto.Categories {
		categoryOptions[i] = _notionNewTableReqSelectOption{
			Name:  formatSelectOption(string(category)),
			Color: colors[i%len(colors)],
		}
	}
	categoryOptions[len(dto.Categories)] = _notionNewTableReqSelectOption{
		Name:  string(sheet.CategoryUnknown),
		Color: Gray,
	}

	requestData := _notionNewTableReq{
		Parent: _notionNewTableReqParent{
			Type:   "page_id",
			PageID: conn.pageID,
		},
		Icon: _notionNewTableReqIcon{
			Type:  "emoji",
			Emoji: "💸",
		},
		Title: []_notionNewTableReqTitle{
			{
				Type: "text",
				Text: _notionNewTableReqText{
					Content: dto.Title,
				},
			},
		},
		Properties: _notionNewTableReqProperties{
			Name: _notionNewTableReqName{
				Title: struct{}{},
			},
			Description: _notionNewTableReqDescription{
				RichText: struct{}{},
			},
			Category: _notionNewTableReqCategory{
				Select: _notionNewTableReqSelect{
					Options: categoryOptions,
				},
			},
			Amount: _notionNewTableReqAmount{
				Number: _notionNewTableReqNumber{
					Format: "real",
				},
			},
			PaymentMethod: _notionNewTableReqCategory{
				Select: _notionNewTableReqSelect{
					Options: []_notionNewTableReqSelectOption{
						{Name: "BOLETO", Color: Yellow},
						{Name: "PIX", Color: Blue},
						{Name: "TED", Color: Green},
						{Name: "CREDIT CARD", Color: Purple},
					},
				},
			},
			Date: _notionNewTableReqDate{
				Date: struct{}{},
			},
		},
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
