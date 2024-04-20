package notionapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type CreateSpendingDatabaseResponse struct {
	ID          string    `json:"id"`
	CreatedTime time.Time `json:"created_time"`
	URL         string    `json:"url"`
}

func (c *Client) CreateSpendingDatabase(
	date time.Time,
) (*CreateSpendingDatabaseResponse, error) {
	url := c.BaseURL

	url.Path = "/v1/databases"

	monthMapper := map[time.Month]string{
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

	year, month, _ := date.Date()

	strMonth := monthMapper[month]

	payload := strings.NewReader(
		fmt.Sprintf(
			"{\"parent\":{\"type\":\"page_id\",\"page_id\":\"aef9e619-15da-4ab2-acd8-5ec9b26fa6e0\"},\"icon\":{\"type\":\"emoji\",\"emoji\":\"💸\"},\"title\":[{\"type\":\"text\",\"text\":{\"content\":\"%s %d\",\"link\":null}}],\"properties\":{\"Name\":{\"title\":{}},\"Description\":{\"select\":{\"options\":[{\"name\":\"Aplicação RDB\",\"color\":\"yellow\"},{\"name\":\"Pagamento de fatura\",\"color\":\"yellow\"},{\"name\":\"Pagamento efetuado\",\"color\":\"yellow\"},{\"name\":\"Resgate RDB\",\"color\":\"yellow\"},{\"name\":\"Transferência enviada\",\"color\":\"yellow\"},{\"name\":\"Transferência Recebida\",\"color\":\"yellow\"}]}},\"Category\":{\"select\":{\"options\":[{\"name\":\"Clothing\",\"color\":\"yellow\"},{\"name\":\"Electricity\",\"color\":\"yellow\"},{\"name\":\"Electronics\",\"color\":\"yellow\"},{\"name\":\"Food and drinks\",\"color\":\"yellow\"},{\"name\":\"Gyms and fitness centers\",\"color\":\"yellow\"},{\"name\":\"Hospital clinics and labs\",\"color\":\"yellow\"},{\"name\":\"Houseware\",\"color\":\"yellow\"},{\"name\":\"Housing\",\"color\":\"yellow\"},{\"name\":\"Investments\",\"color\":\"yellow\"},{\"name\":\"Landmarks and museums\",\"color\":\"yellow\"},{\"name\":\"Pharmacy\",\"color\":\"yellow\"},{\"name\":\"Services\",\"color\":\"yellow\"},{\"name\":\"Telecommunications\",\"color\":\"yellow\"},{\"name\":\"Transfers\",\"color\":\"yellow\"},{\"name\":\"University\",\"color\":\"yellow\"}]}},\"Amount\":{\"number\":{\"format\":\"real\"}},\"Payment Method\":{\"select\":{\"options\":[{\"name\":\"BOLETO\",\"color\":\"yellow\"},{\"name\":\"PIX\",\"color\":\"yellow\"},{\"name\":\"TED\",\"color\":\"yellow\"}]}},\"Date\":{\"date\":{}}}}",
			strMonth,
			year,
		),
	)

	req, err := http.NewRequest("POST", url.String(), payload)
	if err != nil {
		return nil, err
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")
	req.Header.Add("Authorization", "Bearer "+c.Token)
	req.Header.Add("Notion-Version", "2022-06-28")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, parseResError(res)
	}

	decoder := json.NewDecoder(res.Body)
	data := &CreateSpendingDatabaseResponse{}
	if err := decoder.Decode(&data); err != nil {
		return nil, err
	}

	return data, nil

}
