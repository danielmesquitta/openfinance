package sheet

import (
	"fmt"
	"slices"
)

type CreateTableOptions struct {
	Icon    string
	Columns []Column
}

type CreateTableOption func(*CreateTableOptions)

func WithIcon(icon string) CreateTableOption {
	return func(options *CreateTableOptions) {
		options.Icon = icon
	}
}

func WithColumns(columns ...Column) CreateTableOption {
	configuredColumns := slices.Clone(columns)

	return func(options *CreateTableOptions) {
		options.Columns = slices.Clone(configuredColumns)
	}
}

func (options CreateTableOptions) Validate() error {
	for columnIndex, column := range options.Columns {
		if !column.Type().isValid() {
			return fmt.Errorf("column %d has unsupported type %q", columnIndex, column.Type())
		}
		if column.Name() == "" {
			return fmt.Errorf("column %d: name is required", columnIndex)
		}
		if column.Type() != ColumnTypeSelect {
			continue
		}

		for optionIndex, option := range column.SelectOptions() {
			if option.Name() == "" {
				return fmt.Errorf(
					"column %q: select option %d: name is required",
					column.Name(),
					optionIndex,
				)
			}
		}
	}

	return nil
}
