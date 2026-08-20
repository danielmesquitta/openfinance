package sheet

import (
	"errors"
	"fmt"
	"slices"

	"github.com/danielmesquitta/openfinance-to-sheets/internal/domain/entity"
)

type columnCell interface {
	TitleCell | TextCell | NumberCell | SelectCell | DateCell
}

type Column interface {
	Definition() ColumnDefinition
}

type column[T columnCell] struct {
	definition ColumnDefinition
}

func newColumn[T columnCell](name string) column[T] {
	return column[T]{definition: ColumnDefinition{
		name:       name,
		columnType: columnType[T](),
	}}
}

func (c column[T]) Definition() ColumnDefinition {
	definition := c.definition
	definition.selectOptions = slices.Clone(c.definition.selectOptions)

	return definition
}

type TitleColumn struct {
	column[TitleCell]
}

func NewTitleColumn(name string) TitleColumn {
	return TitleColumn{column: newColumn[TitleCell](name)}
}

type TextColumn struct {
	column[TextCell]
}

func NewTextColumn(name string) TextColumn {
	return TextColumn{column: newColumn[TextCell](name)}
}

type NumberColumn struct {
	column[NumberCell]
}

func NewNumberColumn(name string) NumberColumn {
	return NumberColumn{column: newColumn[NumberCell](name)}
}

func (c NumberColumn) Currency(currency Currency) NumberColumn {
	c.definition.currency = currency

	return c
}

type SelectColumn struct {
	column[SelectCell]
}

func NewSelectColumn(name string) SelectColumn {
	return SelectColumn{column: newColumn[SelectCell](name)}
}

func (c SelectColumn) Options(options ...SelectOption) SelectColumn {
	c.definition.selectOptions = make([]SelectOptionDefinition, 0, len(options))
	for _, option := range options {
		c.definition.selectOptions = append(c.definition.selectOptions, option.definition)
	}

	return c
}

type DateColumn struct {
	column[DateCell]
}

func NewDateColumn(name string) DateColumn {
	return DateColumn{column: newColumn[DateCell](name)}
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
	if !d.Type().isValid() {
		return fmt.Errorf("unsupported type %q", d.Type())
	}
	if d.Name() == "" {
		return errors.New("name is required")
	}
	if d.Type() != ColumnTypeSelect {
		return nil
	}

	for optionIndex, option := range d.SelectOptions() {
		if option.Name() == "" {
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

func columnType[T columnCell]() ColumnType {
	var cell T
	switch any(cell).(type) {
	case TitleCell:
		return ColumnTypeTitle
	case TextCell:
		return ColumnTypeText
	case NumberCell:
		return ColumnTypeNumber
	case SelectCell:
		return ColumnTypeSelect
	case DateCell:
		return ColumnTypeDate
	default:
		return ""
	}
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
