package sheet

import "github.com/danielmesquitta/openfinance/internal/domain/entity"

type Table struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type Category string

const (
	CategoryUnknown Category = "Others"
)

type CreateTransactionsTableDTO struct {
	Title      string
	Categories []Category
}

type Provider interface {
	CreateTransactionsTable(
		userID string,
		dto CreateTransactionsTableDTO,
	) (*Table, error)
	InsertTransaction(
		userID string,
		tableID string,
		transaction entity.Transaction,
	) (*Table, error)
}
