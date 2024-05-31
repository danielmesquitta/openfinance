package openfinance

import "github.com/danielmesquitta/openfinance/internal/domain/entity"

type OpenFinanceAPI interface {
	ListTransactions() ([]entity.Transaction, error)
}
