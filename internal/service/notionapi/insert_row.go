package notionapi

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"

	"github.com/danielmesquitta/openfinance/internal/service/meupluggyapi"
)

type InsertRowRes struct {
	ID          string    `json:"id"`
	CreatedTime time.Time `json:"created_time"`
	URL         string    `json:"url"`
}

type InsertRowReq struct {
	Parent     InsertRowReqParent     `json:"parent"`
	Properties InsertRowReqProperties `json:"properties"`
}

type InsertRowReqParent struct {
	DatabaseID string `json:"database_id"`
}

type InsertRowReqProperties struct {
	Name          InsertRowReqName        `json:"Name"`
	Description   InsertRowReqDescription `json:"Description"`
	Category      InsertRowReqSelector    `json:"Category"`
	Amount        InsertRowReqNumber      `json:"Amount"`
	PaymentMethod InsertRowReqSelector    `json:"Payment Method"`
	Date          InsertRowReqDate        `json:"Date"`
}

type InsertRowReqNumber struct {
	Number float64 `json:"number"`
}

type InsertRowReqSelector struct {
	Select InsertRowReqSelect `json:"select"`
}

type InsertRowReqSelect struct {
	Name string `json:"name"`
}

type InsertRowReqDate struct {
	Date InsertRowReqSubDate `json:"date"`
}

type InsertRowReqSubDate struct {
	Start string `json:"start"`
}

type InsertRowReqDescription struct {
	RichText []InsertRowReqRichText `json:"rich_text"`
}

type InsertRowReqRichText struct {
	Text InsertRowReqText `json:"text"`
}

type InsertRowReqText struct {
	Content string `json:"content"`
}

type InsertRowReqName struct {
	Title []InsertRowReqRichText `json:"title"`
}

type InsertRowDTO struct {
	DatabaseID    string
	Name          string
	Description   string
	Category      string
	Amount        float64
	PaymentMethod meupluggyapi.PaymentMethod
	Date          time.Time
}

func (c *Client) InsertRow(dto InsertRowDTO) (*InsertRowRes, error) {
	url := c.BaseURL
	url.Path = "/v1/pages"

	requestData := InsertRowReq{
		Parent: InsertRowReqParent{
			DatabaseID: dto.DatabaseID,
		},
		Properties: InsertRowReqProperties{
			Name: InsertRowReqName{
				Title: []InsertRowReqRichText{
					{
						Text: InsertRowReqText{
							Content: dto.Name,
						},
					},
				},
			},
			Description: InsertRowReqDescription{
				RichText: []InsertRowReqRichText{
					{
						Text: InsertRowReqText{
							Content: dto.Description,
						},
					},
				},
			},
			Category: InsertRowReqSelector{
				Select: InsertRowReqSelect{
					Name: "Others",
				},
			},
			Amount: InsertRowReqNumber{
				Number: dto.Amount,
			},
			PaymentMethod: InsertRowReqSelector{
				Select: InsertRowReqSelect{
					Name: string(dto.PaymentMethod),
				},
			},
			Date: InsertRowReqDate{
				Date: InsertRowReqSubDate{
					Start: dto.Date.Format("2006-01-02T15:04:05.000Z"),
				},
			},
		},
	}

	if dto.Category != "" {
		requestData.Properties.Category = InsertRowReqSelector{
			Select: InsertRowReqSelect{
				Name: formatSelectOption(dto.Category),
			},
		}
	}

	jsonRequestData, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}
	body := bytes.NewReader(jsonRequestData)

	req, err := http.NewRequest("POST", url.String(), body)
	if err != nil {
		return nil, err
	}

	setHeaders(req, c.Token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, parseResError(res)
	}

	decoder := json.NewDecoder(res.Body)
	data := &InsertRowRes{}
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}
