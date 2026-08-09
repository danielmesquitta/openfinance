package sheet

import (
	"context"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

type Provider interface {
	CreateTransactionsTable(ctx context.Context, syncProfileID, title string) (entity.Table, error)
	InsertTransaction(ctx context.Context, syncProfileID, tableID string, transaction entity.Transaction) error
	ListTables(ctx context.Context, syncProfileID string) ([]entity.Table, error)
	ListTransactions(ctx context.Context, syncProfileID, tableID string) ([]entity.Transaction, error)
}
