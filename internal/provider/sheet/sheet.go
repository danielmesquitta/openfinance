package sheet

import (
	"context"
	"time"
)

type Provider interface {
	CreateTable(
		ctx context.Context,
		connectionID string,
		definition TableDefinition,
	) (Table, error)
	EnsureTableColumns(
		ctx context.Context,
		connectionID string,
		tableID string,
		columns ...Column,
	) error
	InsertRow(ctx context.Context, connectionID, tableID string, row Row) error
	ListTables(ctx context.Context, connectionID string) ([]Table, error)
	ListRows(ctx context.Context, connectionID, tableID string) ([]Row, error)
}

type Table struct {
	ID    string
	Title string
}

type Row map[string]Cell

type Cell interface {
	isCell()
}

type TitleCell string
type TextCell string
type NumberCell float64
type SelectCell string
type DateCell time.Time

func (TitleCell) isCell()  {}
func (TextCell) isCell()   {}
func (NumberCell) isCell() {}
func (SelectCell) isCell() {}
func (DateCell) isCell()   {}
