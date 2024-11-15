package brasilapi

import (
	"encoding/json"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/domain/errs"
)

func (c *Client) GetCompanyByID(id string) (entity.Company, error) {
	res, err := c.R().Get("/api/cnpj/v1/" + id)

	if err != nil {
		return entity.Company{}, errs.New(err)
	}

	body := res.Body()
	if statusCode := res.StatusCode(); statusCode < 200 || statusCode >= 300 {
		return entity.Company{}, errs.New(body)
	}

	data := entity.Company{}
	if err := json.Unmarshal(body, &data); err != nil {
		return entity.Company{}, errs.New(err)
	}

	return data, nil
}
