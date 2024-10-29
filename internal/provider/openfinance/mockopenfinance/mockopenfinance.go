package mockopenfinance

import (
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/openfinance"
)

type MockOpenFinance struct{}

func NewMockOpenFinance() *MockOpenFinance {
	return &MockOpenFinance{}
}

func (m MockOpenFinance) ListTransactionsByUserID(
	userID string,
	from, to time.Time,
) ([]entity.Transaction, error) {
	transactions := []entity.Transaction{
		{
			Name:          "99app *99app",
			Amount:        11.70,
			PaymentMethod: entity.PaymentMethodCreditCard,
			Date:          time.Now(),
		},
		{
			Name:          "Uber *Uber *Trip",
			Amount:        10.90,
			PaymentMethod: entity.PaymentMethodCreditCard,
			Date:          time.Now(),
		},
		{
			Name: "Uber *Uber *Trip",
		},
	}

	return transactions, nil
}

var _ openfinance.APIProvider = (*MockOpenFinance)(nil)
