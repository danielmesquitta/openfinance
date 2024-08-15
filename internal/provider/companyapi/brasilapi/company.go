package brasilapi

import (
	"encoding/json"
	"net/http"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

func (c *Client) GetCompanyByID(id string) (entity.Company, error) {
	url := c.BaseURL.String() + "/api/cnpj/v1/" + id

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return entity.Company{}, entity.NewErr(err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return entity.Company{}, entity.NewErr(err)
	}
	if res == nil {
		return entity.Company{}, entity.NewErr("response is nil")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return entity.Company{}, parseResError(res)
	}

	decoder := json.NewDecoder(res.Body)
	data := entity.Company{}
	if err := decoder.Decode(&data); err != nil {
		return entity.Company{}, entity.NewErr(err)
	}

	return data, nil
}
