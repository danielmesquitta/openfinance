package companyapi

import "github.com/danielmesquitta/openfinance/internal/domain/entity"

type APIProvider interface {
	GetCompanyByID(id string) (entity.Company, error)
}
