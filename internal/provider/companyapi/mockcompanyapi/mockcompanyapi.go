package mockcompanyapi

import (
	"math/rand/v2"

	"github.com/go-faker/faker/v4"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/companyapi"
)

const (
	hasTradingNameProbability = 0.5
)

type MockCompanyAPI struct{}

func NewMockCompanyAPI() *MockCompanyAPI {
	return &MockCompanyAPI{}
}

func (m MockCompanyAPI) GetCompanyByID(id string) (entity.Company, error) {
	tradingName := ""
	if rand.Float32() < hasTradingNameProbability {
		tradingName = faker.Word()
	}

	return entity.Company{
		ID:          id,
		Name:        faker.FirstName() + faker.LastName() + " S.A.",
		TradingName: tradingName,
	}, nil
}

var _ companyapi.APIProvider = (*MockCompanyAPI)(nil)
