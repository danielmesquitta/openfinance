package sheet

import (
	"context"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

type Provider interface {
	CreateTransactionsTable(ctx context.Context, userID, title string) (entity.Table, error)
	InsertTransaction(ctx context.Context, userID, tableID string, transaction entity.Transaction) error
	ListTables(ctx context.Context, userID string) ([]entity.Table, error)
	ListTransactions(ctx context.Context, userID, tableID string) ([]entity.Transaction, error)
}
