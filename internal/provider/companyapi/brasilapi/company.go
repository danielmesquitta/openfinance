package brasilapi

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/companyapi"
)

func (c *Client) GetCompanyByID(id string) (entity.Company, error) {
	url := c.BaseURL.String() + "/api/cnpj/v1/" + id

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return entity.Company{}, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return entity.Company{}, fmt.Errorf("error sending request: %w", err)
	}
	if res == nil {
		return entity.Company{}, fmt.Errorf("response is nil")
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return entity.Company{}, parseResError(res)
	}

	decoder := json.NewDecoder(res.Body)
	data := entity.Company{}
	if err := decoder.Decode(&data); err != nil {
		return entity.Company{}, fmt.Errorf("error decoding response: %w", err)
	}

	return data, nil
}

var _ companyapi.API = (*Client)(nil)
