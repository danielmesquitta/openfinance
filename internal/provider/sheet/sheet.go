package sheet

import (
	"context"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

type Table struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type Provider interface {
	CreateTransactionsTable(
		ctx context.Context,
		userID string,
		title string,
	) (*Table, error)
	InsertTransaction(
		ctx context.Context,
		userID string,
		tableID string,
		transaction entity.Transaction,
	) (*Table, error)
}
