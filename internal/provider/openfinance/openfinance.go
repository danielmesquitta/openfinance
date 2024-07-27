package openfinance

import "github.com/danielmesquitta/openfinance/internal/domain/entity"

type API interface {
	ListTransactions() ([]entity.Transaction, error)
}
