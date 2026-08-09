package companyapi

import (
	"context"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

type APIProvider interface {
	GetCompanyByID(ctx context.Context, id string) (entity.Company, error)
}
