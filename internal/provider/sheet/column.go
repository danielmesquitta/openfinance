package sheet

import (
	"slices"

	"github.com/danielmesquitta/openfinance-to-sheets/internal/domain/entity"
)

type Column struct {
	name          string
	columnType    ColumnType
	currency      string
	selectOptions []SelectOption
}

type ColumnType string

const (
	ColumnTypeTitle  ColumnType = "title"
	ColumnTypeText   ColumnType = "text"
	ColumnTypeNumber ColumnType = "number"
	ColumnTypeSelect ColumnType = "select"
	ColumnTypeDate   ColumnType = "date"
)

func NewTitleColumn(name string) Column {
	return Column{name: name, columnType: ColumnTypeTitle}
}

func NewTextColumn(name string) Column {
	return Column{name: name, columnType: ColumnTypeText}
}

func NewDateColumn(name string) Column {
	return Column{name: name, columnType: ColumnTypeDate}
}

type NumberColumnOptions struct {
	Currency string
}

type NumberColumnOption func(*NumberColumnOptions)

func NewNumberColumn(name string, options ...NumberColumnOption) Column {
	numberOptions := NumberColumnOptions{}
	for _, option := range options {
		if option != nil {
			option(&numberOptions)
		}
	}

	return Column{
		name:       name,
		columnType: ColumnTypeNumber,
		currency:   numberOptions.Currency,
	}
}

func WithCurrency(currency string) NumberColumnOption {
	return func(options *NumberColumnOptions) {
		options.Currency = currency
	}
}

type SelectColumnOptions struct {
	Options []SelectOption
}

type SelectColumnOption func(*SelectColumnOptions)

func NewSelectColumn(name string, options ...SelectColumnOption) Column {
	selectOptions := SelectColumnOptions{}
	for _, option := range options {
		if option != nil {
			option(&selectOptions)
		}
	}

	return Column{
		name:          name,
		columnType:    ColumnTypeSelect,
		selectOptions: slices.Clone(selectOptions.Options),
	}
}

func WithSelectOptions(options ...SelectOption) SelectColumnOption {
	configuredOptions := slices.Clone(options)

	return func(columnOptions *SelectColumnOptions) {
		columnOptions.Options = slices.Clone(configuredOptions)
	}
}

func (c Column) Name() string {
	return c.name
}

func (c Column) Type() ColumnType {
	return c.columnType
}

func (c Column) Currency() string {
	return c.currency
}

func (c Column) SelectOptions() []SelectOption {
	return slices.Clone(c.selectOptions)
}

func (t ColumnType) isValid() bool {
	switch t {
	case ColumnTypeTitle,
		ColumnTypeText,
		ColumnTypeNumber,
		ColumnTypeSelect,
		ColumnTypeDate:
		return true
	default:
		return false
	}
}

type SelectOption struct {
	name  string
	color entity.Color
}

type SelectOptionOptions struct {
	Color entity.Color
}

type SelectOptionOption func(*SelectOptionOptions)

func NewSelectOption(name string, options ...SelectOptionOption) SelectOption {
	selectOptionOptions := SelectOptionOptions{}
	for _, option := range options {
		if option != nil {
			option(&selectOptionOptions)
		}
	}

	return SelectOption{name: name, color: selectOptionOptions.Color}
}

func WithColor(color entity.Color) SelectOptionOption {
	return func(options *SelectOptionOptions) {
		options.Color = color
	}
}

func (o SelectOption) Name() string {
	return o.name
}

func (o SelectOption) Color() entity.Color {
	return o.color
}
