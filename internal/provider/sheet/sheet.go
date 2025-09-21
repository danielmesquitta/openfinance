package sheet

import (
	"context"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

type Table struct {
	ID       string  `json:"id,omitzero"`
	Title    *string `json:"title,omitzero"`
	Archived bool    `json:"archived,omitzero"`
	InTrash  bool    `json:"in_trash,omitzero"`
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
	ListTables(
		ctx context.Context,
		userID string,
	) ([]Table, error)
	ListTransactions(
		ctx context.Context,
		userID string,
		tableID string,
	) ([]entity.Transaction, error)
}
