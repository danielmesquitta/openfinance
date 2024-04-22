package notionapi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var monthMapper = map[time.Month]string{
	1:  "Jan",
	2:  "Feb",
	3:  "Mar",
	4:  "Apr",
	5:  "May",
	6:  "Jun",
	7:  "Jul",
	8:  "Aug",
	9:  "Sep",
	10: "Oct",
	11: "Nov",
	12: "Dec",
}

type CreateDBRes struct {
	ID          string    `json:"id"`
	CreatedTime time.Time `json:"created_time"`
	URL         string    `json:"url"`
}

type CreateDBReq struct {
	Parent     CreateDBReqParent     `json:"parent"`
	Icon       CreateDBReqIcon       `json:"icon"`
	Title      []CreateDBReqTitle    `json:"title"`
	Properties CreateDBReqProperties `json:"properties"`
}

type CreateDBReqIcon struct {
	Type  string `json:"type"`
	Emoji string `json:"emoji"`
}

type CreateDBReqParent struct {
	Type   string `json:"type"`
	PageID string `json:"page_id"`
}

type CreateDBReqProperties struct {
	Name          CreateDBReqName        `json:"Name"`
	Description   CreateDBReqDescription `json:"Description"`
	Category      CreateDBReqCategory    `json:"Category"`
	Amount        CreateDBReqAmount      `json:"Amount"`
	PaymentMethod CreateDBReqCategory    `json:"Payment Method"`
	Date          CreateDBReqDate        `json:"Date"`
}

type CreateDBReqAmount struct {
	Number CreateDBReqNumber `json:"number"`
}

type CreateDBReqNumber struct {
	Format string `json:"format"`
}

type CreateDBReqCategory struct {
	Select CreateDBReqSelect `json:"select"`
}

type CreateDBReqSelect struct {
	Options []CreateDBReqSelectOption `json:"options"`
}

type CreateDBReqSelectOption struct {
	Name  string `json:"name"`
	Color Color  `json:"color"`
}

type CreateDBReqDate struct {
	Date struct{} `json:"date"`
}

type CreateDBReqDescription struct {
	RichText struct{} `json:"rich_text"`
}

type CreateDBReqName struct {
	Title struct{} `json:"title"`
}

type CreateDBReqTitle struct {
	Type string          `json:"type"`
	Text CreateDBReqText `json:"text"`
}

type CreateDBReqText struct {
	Content string `json:"content"`
}

type CreateDBDTO struct {
	PageID     string
	Date       time.Time
	Categories []string
}

func (c *Client) CreateDB(
	dto CreateDBDTO,
) (*CreateDBRes, error) {
	url := c.BaseURL
	url.Path = "/v1/databases"

	categoryOptions := make([]CreateDBReqSelectOption, len(dto.Categories)+1)
	for i, category := range dto.Categories {
		categoryOptions[i] = CreateDBReqSelectOption{
			Name:  formatSelectOption(category),
			Color: colors[i%len(colors)],
		}
	}
	categoryOptions[len(dto.Categories)] = CreateDBReqSelectOption{
		Name:  "Others",
		Color: Gray,
	}

	year, month, _ := dto.Date.Date()
	strMonth := monthMapper[month]

	requestData := CreateDBReq{
		Parent: CreateDBReqParent{
			Type:   "page_id",
			PageID: dto.PageID,
		},
		Icon: CreateDBReqIcon{
			Type:  "emoji",
			Emoji: "💸",
		},
		Title: []CreateDBReqTitle{
			{
				Type: "text",
				Text: CreateDBReqText{
					Content: fmt.Sprintf("%s %d", strMonth, year),
				},
			},
		},
		Properties: CreateDBReqProperties{
			Name: CreateDBReqName{
				Title: struct{}{},
			},
			Description: CreateDBReqDescription{
				RichText: struct{}{},
			},
			Category: CreateDBReqCategory{
				Select: CreateDBReqSelect{
					Options: categoryOptions,
				},
			},
			Amount: CreateDBReqAmount{
				Number: CreateDBReqNumber{
					Format: "real",
				},
			},
			PaymentMethod: CreateDBReqCategory{
				Select: CreateDBReqSelect{
					Options: []CreateDBReqSelectOption{
						{Name: "BOLETO", Color: Yellow},
						{Name: "PIX", Color: Blue},
						{Name: "TED", Color: Green},
						{Name: "CREDIT CARD", Color: Purple},
					},
				},
			},
			Date: CreateDBReqDate{
				Date: struct{}{},
			},
		},
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
	data := &CreateDBRes{}
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	return data, nil
}
