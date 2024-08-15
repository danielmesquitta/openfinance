package companyapi

import "github.com/danielmesquitta/openfinance/internal/domain/entity"

type API interface {
	GetCompanyByID(id string) (entity.Company, error)
}
