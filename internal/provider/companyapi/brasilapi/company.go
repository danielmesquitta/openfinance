package brasilapi

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

type companyResponse struct {
	ID          string `json:"cnpj"`
	Name        string `json:"razao_social"`
	TradingName string `json:"nome_fantasia"`
}

// GetCompanyByID gets a company by id.
func (c *Client) GetCompanyByID(ctx context.Context, id string) (entity.Company, error) {
	res, err := c.R().SetContext(ctx).Get("/api/cnpj/v1/" + id)
	if err != nil {
		return entity.Company{}, fmt.Errorf("failed to get company by id %s: %w", id, err)
	}

	body := res.Body()
	if res.IsError() {
		return entity.Company{}, fmt.Errorf(
			"failed to get company by id %s with response %s",
			id,
			body,
		)
	}

	data := companyResponse{}
	if err := json.Unmarshal(body, &data); err != nil {
		return entity.Company{}, fmt.Errorf(
			"failed to unmarshal company by id %s: %w",
			id,
			err,
		)
	}

	return entity.Company{
		ID:          data.ID,
		Name:        data.Name,
		TradingName: data.TradingName,
	}, nil
}
