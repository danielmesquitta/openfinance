package notionapi

import (
	"net/http"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
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
	Name          notionNewTableReqName        `json:"Name"`
	Description   notionNewTableReqDescription `json:"Description"`
	Category      notionNewTableReqCategory    `json:"Category"`
	Amount        notionNewTableReqAmount      `json:"Amount"`
	PaymentMethod notionNewTableReqCategory    `json:"Payment Method"`
	Date          notionNewTableReqDate        `json:"Date"`
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

type notionNewTableReqDescription struct {
	RichText struct{} `json:"rich_text"`
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
			Description: notionNewTableReqDescription{
				RichText: struct{}{},
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
