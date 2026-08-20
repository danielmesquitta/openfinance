package sheet

import (
	"reflect"
	"strings"
	"testing"

	"github.com/danielmesquitta/openfinance-to-sheets/internal/domain/entity"
)

func TestColumnTypesAreDerivedFromCells(t *testing.T) {
	tests := []struct {
		name       string
		definition ColumnDefinition
		columnType ColumnType
	}{
		{name: "title", definition: NewTitleColumn("Name").Definition(), columnType: ColumnTypeTitle},
		{name: "text", definition: NewTextColumn("Notes").Definition(), columnType: ColumnTypeText},
		{name: "number", definition: NewNumberColumn("Amount").Definition(), columnType: ColumnTypeNumber},
		{name: "select", definition: NewSelectColumn("Category").Definition(), columnType: ColumnTypeSelect},
		{name: "date", definition: NewDateColumn("Date").Definition(), columnType: ColumnTypeDate},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.definition.Name() == "" || test.definition.Type() != test.columnType {
				t.Fatalf("column definition = %#v", test.definition)
			}
			if test.definition.Currency() != "" || len(test.definition.SelectOptions()) != 0 {
				t.Fatalf("unexpected optional configuration: %#v", test.definition)
			}
		})
	}
}

func TestColumnMethodSetsExposeOnlyCompatibleConfiguration(t *testing.T) {
	tests := []struct {
		name         string
		column       any
		wantCurrency bool
		wantOptions  bool
	}{
		{name: "title", column: NewTitleColumn("Title")},
		{name: "text", column: NewTextColumn("Text")},
		{name: "number", column: NewNumberColumn("Number"), wantCurrency: true},
		{name: "select", column: NewSelectColumn("Select"), wantOptions: true},
		{name: "date", column: NewDateColumn("Date")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			columnType := reflect.TypeOf(test.column)
			_, hasCurrency := columnType.MethodByName("Currency")
			_, hasOptions := columnType.MethodByName("Options")
			if hasCurrency != test.wantCurrency || hasOptions != test.wantOptions {
				t.Fatalf(
					"method set Currency=%t Options=%t, want Currency=%t Options=%t",
					hasCurrency,
					hasOptions,
					test.wantCurrency,
					test.wantOptions,
				)
			}
		})
	}
}

func TestColumnChainsUseLastValue(t *testing.T) {
	number := NewNumberColumn("Amount").
		Currency("USD").
		Currency("BRL").
		Definition()
	if number.Currency() != Currency("BRL") {
		t.Fatalf("currency = %q, want BRL", number.Currency())
	}

	first := NewSelectOption("First").Color(entity.Blue)
	second := NewSelectOption("Second").Color(entity.Red)
	selectColumn := NewSelectColumn("Category").
		Options(first).
		Options(second).
		Definition()
	options := selectColumn.SelectOptions()
	if len(options) != 1 || options[0].Name() != "Second" || options[0].Color() != entity.Red {
		t.Fatalf("select options = %#v", options)
	}
}

func TestFluentDefinitionsAreImmutable(t *testing.T) {
	configuredOptions := []SelectOption{NewSelectOption("Food").Color(entity.Red)}
	column := NewSelectColumn("Category").Options(configuredOptions...)
	table := NewTable("Transactions").
		SetIcon("first").
		SetIcon("💸").
		AddColumn(NewTitleColumn("Name")).
		AddColumn(column)

	configuredOptions[0] = NewSelectOption("Changed")
	column = column.Options(NewSelectOption("Also changed"))
	tableColumns := table.Columns()
	tableColumns[0] = NewTextColumn("Changed").Definition()
	returnedOptions := table.Columns()[1].SelectOptions()
	returnedOptions[0] = SelectOptionDefinition{name: "Mutated"}

	storedColumns := table.Columns()
	if table.Icon() != "💸" || len(storedColumns) != 2 ||
		storedColumns[0].Type() != ColumnTypeTitle ||
		storedColumns[1].SelectOptions()[0].Name() != "Food" {
		t.Fatalf("table definition = %#v", table)
	}
}

func TestTableDefinitionValidate(t *testing.T) {
	var nilColumn Column
	tests := []struct {
		name        string
		definition  TableDefinition
		wantErrPart string
	}{
		{
			name:        "nil column",
			definition:  NewTable("Table").AddColumn(nilColumn),
			wantErrPart: "unsupported type",
		},
		{
			name:        "title name",
			definition:  NewTable("Table").AddColumn(NewTitleColumn("")),
			wantErrPart: "name is required",
		},
		{
			name:        "text name",
			definition:  NewTable("Table").AddColumn(NewTextColumn("")),
			wantErrPart: "name is required",
		},
		{
			name:        "number name",
			definition:  NewTable("Table").AddColumn(NewNumberColumn("")),
			wantErrPart: "name is required",
		},
		{
			name:        "select name",
			definition:  NewTable("Table").AddColumn(NewSelectColumn("")),
			wantErrPart: "name is required",
		},
		{
			name:        "date name",
			definition:  NewTable("Table").AddColumn(NewDateColumn("")),
			wantErrPart: "name is required",
		},
		{
			name: "select option name",
			definition: NewTable("Table").AddColumn(
				NewSelectColumn("Category").Options(NewSelectOption("")),
			),
			wantErrPart: "select option 0: name is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.definition.Validate()
			if err == nil || !strings.Contains(err.Error(), test.wantErrPart) {
				t.Fatalf("Validate() error = %v, want containing %q", err, test.wantErrPart)
			}
		})
	}

	valid := NewTable("Table").
		AddColumn(NewTitleColumn("Name")).
		AddColumn(NewNumberColumn("Amount")).
		AddColumn(NewSelectColumn("Category"))
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() valid definition error = %v", err)
	}
	if err := (TableDefinition{}).Validate(); err != nil {
		t.Fatalf("Validate() empty definition error = %v", err)
	}
}
