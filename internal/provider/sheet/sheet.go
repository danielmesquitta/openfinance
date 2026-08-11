package sheet

import (
	"context"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

type Provider interface {
	CreateTransactionsTable(ctx context.Context, ingestProfileID, title string) (entity.Table, error)
	InsertTransaction(ctx context.Context, ingestProfileID, tableID string, transaction entity.Transaction) error
	ListTables(ctx context.Context, ingestProfileID string) ([]entity.Table, error)
	ListTransactions(ctx context.Context, ingestProfileID, tableID string) ([]entity.Transaction, error)
}
