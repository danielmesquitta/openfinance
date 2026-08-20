package sheet

import (
	"errors"
	"fmt"
	"slices"

	"github.com/danielmesquitta/openfinance-to-sheets/internal/domain/entity"
)

type Column interface {
	Definition() ColumnDefinition
}

type column struct {
	definition ColumnDefinition
}

func newColumn(name string, columnType ColumnType) column {
	return column{definition: ColumnDefinition{
		name:       name,
		columnType: columnType,
	}}
}

func (c column) Definition() ColumnDefinition {
	definition := c.definition
	definition.selectOptions = slices.Clone(c.definition.selectOptions)

	return definition
}

type TitleColumn struct {
	column
}

func NewTitleColumn(name string) TitleColumn {
	return TitleColumn{column: newColumn(name, ColumnTypeTitle)}
}

type TextColumn struct {
	column
}

func NewTextColumn(name string) TextColumn {
	return TextColumn{column: newColumn(name, ColumnTypeText)}
}

type NumberColumn struct {
	column
}

func NewNumberColumn(name string) NumberColumn {
	return NumberColumn{column: newColumn(name, ColumnTypeNumber)}
}

func (c NumberColumn) Currency(currency Currency) NumberColumn {
	c.definition.currency = currency

	return c
}

type SelectColumn struct {
	column
}

func NewSelectColumn(name string) SelectColumn {
	return SelectColumn{column: newColumn(name, ColumnTypeSelect)}
}

func (c SelectColumn) Options(options ...SelectOption) SelectColumn {
	definitions := make([]SelectOptionDefinition, len(options))
	for index, option := range options {
		definitions[index] = option.definition
	}
	c.definition.selectOptions = definitions

	return c
}

type DateColumn struct {
	column
}

func NewDateColumn(name string) DateColumn {
	return DateColumn{column: newColumn(name, ColumnTypeDate)}
}

type ColumnDefinition struct {
	name          string
	columnType    ColumnType
	currency      Currency
	selectOptions []SelectOptionDefinition
}

func (d ColumnDefinition) Name() string {
	return d.name
}

func (d ColumnDefinition) Type() ColumnType {
	return d.columnType
}

func (d ColumnDefinition) Currency() Currency {
	return d.currency
}

func (d ColumnDefinition) SelectOptions() []SelectOptionDefinition {
	return slices.Clone(d.selectOptions)
}

func (d ColumnDefinition) Validate() error {
	if !d.columnType.isValid() {
		return fmt.Errorf("unsupported type %q", d.columnType)
	}
	if d.name == "" {
		return errors.New("name is required")
	}
	if d.columnType != ColumnTypeSelect {
		return nil
	}

	for optionIndex, option := range d.selectOptions {
		if option.name == "" {
			return fmt.Errorf("select option %d: name is required", optionIndex)
		}
	}

	return nil
}

type ColumnType string

const (
	ColumnTypeTitle  ColumnType = "title"
	ColumnTypeText   ColumnType = "text"
	ColumnTypeNumber ColumnType = "number"
	ColumnTypeSelect ColumnType = "select"
	ColumnTypeDate   ColumnType = "date"
)

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

type Currency string

type SelectOption struct {
	definition SelectOptionDefinition
}

func NewSelectOption(name string) SelectOption {
	return SelectOption{definition: SelectOptionDefinition{name: name}}
}

func (o SelectOption) Color(color entity.Color) SelectOption {
	o.definition.color = color

	return o
}

type SelectOptionDefinition struct {
	name  string
	color entity.Color
}

func (d SelectOptionDefinition) Name() string {
	return d.name
}

func (d SelectOptionDefinition) Color() entity.Color {
	return d.color
}
