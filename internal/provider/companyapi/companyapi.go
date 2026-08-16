package companyapi

import (
	"context"

	"github.com/danielmesquitta/openfinance-to-sheets/internal/domain/entity"
)

type APIProvider interface {
	GetCompanyByID(ctx context.Context, id string) (entity.Company, error)
}
