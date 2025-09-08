package brasilapi

import (
	"encoding/json"
	"fmt"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

// GetCompanyByID gets a company by id.
func (c *Client) GetCompanyByID(id string) (entity.Company, error) {
	res, err := c.R().Get("/api/cnpj/v1/" + id)

	if err != nil {
		return entity.Company{}, fmt.Errorf("failed to get company by id: %w", err)
	}

	body := res.Body()
	if res.IsError() {
		return entity.Company{}, fmt.Errorf("failed to get company by id: %+v", body)
	}

	data := entity.Company{}
	if err := json.Unmarshal(body, &data); err != nil {
		return entity.Company{}, fmt.Errorf("failed to unmarshal company by id: %w", err)
	}

	return data, nil
}
