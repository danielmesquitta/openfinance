package sheet

import (
	"strings"
	"testing"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

func TestColumnConstructors(t *testing.T) {
	tests := []struct {
		name       string
		column     Column
		columnType ColumnType
	}{
		{name: "title", column: NewTitleColumn("Name"), columnType: ColumnTypeTitle},
		{name: "text", column: NewTextColumn("Notes"), columnType: ColumnTypeText},
		{name: "number", column: NewNumberColumn("Amount"), columnType: ColumnTypeNumber},
		{name: "select", column: NewSelectColumn("Category"), columnType: ColumnTypeSelect},
		{name: "date", column: NewDateColumn("Date"), columnType: ColumnTypeDate},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.column.Name() == "" || test.column.Type() != test.columnType {
				t.Fatalf("column = %#v", test.column)
			}
			if test.column.Currency() != "" || len(test.column.SelectOptions()) != 0 {
				t.Fatalf("unexpected optional configuration: %#v", test.column)
			}
		})
	}
}

func TestNumberColumnOptionsAreNilSafeAndLastWins(t *testing.T) {
	column := NewNumberColumn(
		"Amount",
		WithCurrency("USD"),
		nil,
		WithCurrency("BRL"),
	)
	if column.Currency() != "BRL" {
		t.Fatalf("currency = %q, want BRL", column.Currency())
	}
}

func TestSelectColumnAndOptionConfiguration(t *testing.T) {
	option := NewSelectOption(
		"Food",
		WithColor(entity.Blue),
		nil,
		WithColor(entity.Red),
	)
	if option.Name() != "Food" || option.Color() != entity.Red {
		t.Fatalf("select option = %#v", option)
	}
	if color := NewSelectOption("No color").Color(); color != "" {
		t.Fatalf("optional color = %q, want empty", color)
	}

	first := NewSelectOption("First")
	second := NewSelectOption("Second")
	column := NewSelectColumn(
		"Category",
		WithSelectOptions(first),
		nil,
		WithSelectOptions(second),
	)
	options := column.SelectOptions()
	if len(options) != 1 || options[0].Name() != "Second" {
		t.Fatalf("select options = %#v", options)
	}
}

func TestFunctionalOptionsCopySlices(t *testing.T) {
	columns := []Column{NewTitleColumn("Name")}
	tableOption := WithColumns(columns...)
	resolved := resolveTableOptions(WithIcon("first"), nil, WithIcon("last"), tableOption)
	columns[0] = NewTextColumn("Changed")

	if resolved.Icon != "last" || len(resolved.Columns) != 1 ||
		resolved.Columns[0].Type() != ColumnTypeTitle {
		t.Fatalf("resolved table options = %#v", resolved)
	}

	configuredSelectOptions := []SelectOption{NewSelectOption("Food")}
	selectOption := WithSelectOptions(configuredSelectOptions...)
	column := NewSelectColumn("Category", selectOption)
	configuredSelectOptions[0] = NewSelectOption("Changed")
	returnedOptions := column.SelectOptions()
	returnedOptions[0] = NewSelectOption("Also changed")

	if got := column.SelectOptions()[0].Name(); got != "Food" {
		t.Fatalf("stored select option = %q, want Food", got)
	}
}

func TestCreateTableOptionsValidate(t *testing.T) {
	tests := []struct {
		name        string
		columns     []Column
		wantErrPart string
	}{
		{name: "zero column", columns: []Column{{}}, wantErrPart: "unsupported type"},
		{name: "title name", columns: []Column{NewTitleColumn("")}, wantErrPart: "name is required"},
		{name: "text name", columns: []Column{NewTextColumn("")}, wantErrPart: "name is required"},
		{name: "number name", columns: []Column{NewNumberColumn("")}, wantErrPart: "name is required"},
		{name: "select name", columns: []Column{NewSelectColumn("")}, wantErrPart: "name is required"},
		{name: "date name", columns: []Column{NewDateColumn("")}, wantErrPart: "name is required"},
		{
			name: "select option name",
			columns: []Column{NewSelectColumn(
				"Category",
				WithSelectOptions(NewSelectOption("")),
			)},
			wantErrPart: "select option 0: name is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := (CreateTableOptions{Columns: test.columns}).Validate()
			if err == nil || !strings.Contains(err.Error(), test.wantErrPart) {
				t.Fatalf("Validate() error = %v, want containing %q", err, test.wantErrPart)
			}
		})
	}

	valid := CreateTableOptions{Columns: []Column{
		NewTitleColumn("Name"),
		NewNumberColumn("Amount"),
		NewSelectColumn("Category"),
	}}
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() valid options error = %v", err)
	}
	if err := (CreateTableOptions{}).Validate(); err != nil {
		t.Fatalf("Validate() empty options error = %v", err)
	}
}

func resolveTableOptions(options ...CreateTableOption) CreateTableOptions {
	resolved := CreateTableOptions{}
	for _, option := range options {
		if option != nil {
			option(&resolved)
		}
	}

	return resolved
}
