package sheet

import (
	"fmt"
	"slices"
)

type TableDefinition struct {
	title   string
	icon    string
	columns []ColumnDefinition
}

func NewTable(title string) TableDefinition {
	return TableDefinition{title: title}
}

func (d TableDefinition) SetIcon(icon string) TableDefinition {
	d.icon = icon

	return d
}

func (d TableDefinition) AddColumn[C Column](column C) TableDefinition {
	d.columns = slices.Clone(d.columns)
	if any(column) == nil {
		d.columns = append(d.columns, ColumnDefinition{})

		return d
	}

	d.columns = append(d.columns, column.Definition())

	return d
}

func (d TableDefinition) Title() string {
	return d.title
}

func (d TableDefinition) Icon() string {
	return d.icon
}

func (d TableDefinition) Columns() []ColumnDefinition {
	return slices.Clone(d.columns)
}

func (d TableDefinition) Validate() error {
	for columnIndex, column := range d.columns {
		if err := column.Validate(); err != nil {
			return fmt.Errorf("column %d: %w", columnIndex, err)
		}
	}

	return nil
}
