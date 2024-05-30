package sheet

import "github.com/danielmesquitta/openfinance/internal/domain/entity"

type Table struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

type NewTableDTO struct {
	ParentID   string
	Title      string
	Categories []string
}

type SheetProvider interface {
	NewTable(dto NewTableDTO) (*Table, error)
	InsertRow(databaseID string, transaction entity.Transaction) (*Table, error)
}
