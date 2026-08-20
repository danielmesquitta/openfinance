package ingest

import (
	"reflect"
	"testing"
	"time"

	"github.com/danielmesquitta/openfinance-to-sheets/internal/domain/entity"
	"github.com/danielmesquitta/openfinance-to-sheets/internal/provider/sheet"
)

func TestTransactionTableDefinition(t *testing.T) {
	tests := []struct {
		name     string
		language entity.Language
		columns  []sheet.ColumnDefinition
	}{
		{
			name:     "English default",
			language: "",
			columns: []sheet.ColumnDefinition{
				sheet.NewTitleColumn("Name").Definition(),
				sheet.NewSelectColumn("Category").Options(
					sheet.NewSelectOption("Food").Color(entity.Red),
					sheet.NewSelectOption("Others").Color(entity.Gray),
				).Definition(),
				sheet.NewNumberColumn("Amount").
					Currency(sheet.Currency("BRL")).Definition(),
				sheet.NewSelectColumn("Payment Method").Options(
					sheet.NewSelectOption("BOLETO").Color(entity.Yellow),
					sheet.NewSelectOption("PIX").Color(entity.Blue),
					sheet.NewSelectOption("TED").Color(entity.Green),
					sheet.NewSelectOption("CREDIT CARD").Color(entity.Purple),
				).Definition(),
				sheet.NewTextColumn("Card Last Digits").Definition(),
				sheet.NewDateColumn("Date").Definition(),
			},
		},
		{
			name:     "Brazilian Portuguese",
			language: entity.LanguagePortugueseBrazil,
			columns: []sheet.ColumnDefinition{
				sheet.NewTitleColumn("Nome").Definition(),
				sheet.NewSelectColumn("Categoria").Options(
					sheet.NewSelectOption("Food").Color(entity.Red),
					sheet.NewSelectOption("Others").Color(entity.Gray),
				).Definition(),
				sheet.NewNumberColumn("Valor").
					Currency(sheet.Currency("BRL")).Definition(),
				sheet.NewSelectColumn("Forma de pagamento").Options(
					sheet.NewSelectOption("BOLETO").Color(entity.Yellow),
					sheet.NewSelectOption("PIX").Color(entity.Blue),
					sheet.NewSelectOption("TED").Color(entity.Green),
					sheet.NewSelectOption("CARTÃO DE CRÉDITO").Color(entity.Purple),
				).Definition(),
				sheet.NewTextColumn("Últimos dígitos do cartão").Definition(),
				sheet.NewDateColumn("Data").Definition(),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			settings := entity.IngestProfileSettings{
				Language:   test.language,
				Categories: []entity.Category{"Food", "Others"},
				ColorsByCategory: map[entity.Category]entity.Color{
					"Food": entity.Red, "Others": entity.Gray,
				},
			}

			got := transactionTableDefinition("Transactions", settings)
			if got.Title() != "Transactions" || got.Icon() != "💸" ||
				!reflect.DeepEqual(got.Columns(), test.columns) {
				t.Fatalf("definition = %#v, want columns %#v", got, test.columns)
			}
		})
	}
}

func TestTransactionTableDefinitionIncludesLocalizedBudgetGroup(t *testing.T) {
	tests := []struct {
		language   entity.Language
		columnName string
	}{
		{language: entity.LanguageEnglish, columnName: "Budget Group"},
		{language: entity.LanguagePortugueseBrazil, columnName: "Grupo do orçamento"},
	}

	for _, test := range tests {
		t.Run(string(test.language), func(t *testing.T) {
			settings := entity.IngestProfileSettings{
				Language:         test.language,
				Categories:       []entity.Category{"Food"},
				ColorsByCategory: map[entity.Category]entity.Color{"Food": entity.Red},
				BudgetGroups:     []entity.BudgetGroup{"Fixed Costs", "Other"},
				ColorsByBudgetGroup: map[entity.BudgetGroup]entity.Color{
					"Fixed Costs": entity.Red,
					"Other":       entity.Gray,
				},
			}

			columns := transactionTableDefinition("Transactions", settings).Columns()
			if len(columns) != 7 || columns[2].Name() != test.columnName ||
				columns[2].Type() != sheet.ColumnTypeSelect {
				t.Fatalf("columns = %#v", columns)
			}
			options := columns[2].SelectOptions()
			if len(options) != 2 || options[0].Name() != "Fixed Costs" ||
				options[0].Color() != entity.Red || options[1].Name() != "Other" ||
				options[1].Color() != entity.Gray {
				t.Fatalf("budget group options = %#v", options)
			}
		})
	}
}

func TestTransactionRowRoundTrip(t *testing.T) {
	cardLastDigits := "1234"
	transaction := entity.Transaction{
		Name:           "Store",
		Category:       "Food",
		BudgetGroup:    "Lifestyle",
		Amount:         42.5,
		PaymentMethod:  entity.PaymentMethodCreditCard,
		CardLastDigits: &cardLastDigits,
		Date:           time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC),
	}

	tests := []struct {
		name               string
		language           entity.Language
		columns            transactionTableColumns
		paymentMethodLabel string
	}{
		{
			name:               "English",
			language:           entity.LanguageEnglish,
			columns:            transactionTableLocalizationFor(entity.LanguageEnglish).columns,
			paymentMethodLabel: "CREDIT CARD",
		},
		{
			name:               "Brazilian Portuguese",
			language:           entity.LanguagePortugueseBrazil,
			columns:            transactionTableLocalizationFor(entity.LanguagePortugueseBrazil).columns,
			paymentMethodLabel: "CARTÃO DE CRÉDITO",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			row := transactionToRow(transaction, test.language)
			if row[test.columns.name] != sheet.TitleCell("Store") ||
				row[test.columns.category] != sheet.SelectCell("Food") ||
				row[test.columns.budgetGroup] != sheet.SelectCell("Lifestyle") ||
				row[test.columns.amount] != sheet.NumberCell(42.5) ||
				row[test.columns.paymentMethod] != sheet.SelectCell(test.paymentMethodLabel) ||
				row[test.columns.cardLastDigits] != sheet.TextCell("1234") ||
				row[test.columns.date] != sheet.DateCell(transaction.Date) {
				t.Fatalf("row = %#v", row)
			}

			got, err := rowToTransaction(row, test.language)
			if err != nil {
				t.Fatalf("rowToTransaction() error = %v", err)
			}
			if !reflect.DeepEqual(got, transaction) {
				t.Fatalf("transaction = %#v, want %#v", got, transaction)
			}
		})
	}
}

func TestTransactionRowEmptyCardLastDigits(t *testing.T) {
	columns := transactionTableLocalizationFor(entity.LanguagePortugueseBrazil).columns
	row := transactionToRow(entity.Transaction{
		PaymentMethod: entity.PaymentMethodPix,
	}, entity.LanguagePortugueseBrazil)
	if row[columns.cardLastDigits] != sheet.TextCell("") {
		t.Fatalf("card cell = %#v", row[columns.cardLastDigits])
	}

	transaction, err := rowToTransaction(row, entity.LanguagePortugueseBrazil)
	if err != nil {
		t.Fatalf("rowToTransaction() error = %v", err)
	}
	if transaction.CardLastDigits != nil {
		t.Fatalf("card last digits = %#v, want nil", transaction.CardLastDigits)
	}
}

func TestTransactionToRowOmitsEmptyPaymentMethod(t *testing.T) {
	columns := transactionTableLocalizationFor(entity.LanguageEnglish).columns
	row := transactionToRow(entity.Transaction{Name: "Payment made"}, entity.LanguageEnglish)

	if paymentMethod, exists := row[columns.paymentMethod]; exists {
		t.Fatalf("payment method cell = %#v, want omitted", paymentMethod)
	}
	if budgetGroup, exists := row[columns.budgetGroup]; exists {
		t.Fatalf("budget group cell = %#v, want omitted", budgetGroup)
	}

	transaction, err := rowToTransaction(row, entity.LanguageEnglish)
	if err != nil {
		t.Fatalf("rowToTransaction() error = %v", err)
	}
	if transaction.PaymentMethod != "" {
		t.Fatalf("payment method = %q, want empty", transaction.PaymentMethod)
	}
}

func TestRowToTransactionAcceptsLegacyRowWithoutBudgetGroup(t *testing.T) {
	transaction := entity.Transaction{
		Name:     "Store",
		Category: "Food",
		Amount:   10,
		Date:     time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC),
	}
	row := transactionToRow(transaction, entity.LanguageEnglish)
	delete(row, "Budget Group")

	got, err := rowToTransaction(row, entity.LanguageEnglish)
	if err != nil {
		t.Fatalf("rowToTransaction() error = %v", err)
	}
	if !reflect.DeepEqual(got, transaction) {
		t.Fatalf("transaction = %#v, want %#v", got, transaction)
	}
}

func TestRowToTransactionRejectsUnexpectedCellType(t *testing.T) {
	_, err := rowToTransaction(
		sheet.Row{"Valor": sheet.TextCell("42")},
		entity.LanguagePortugueseBrazil,
	)
	if err == nil {
		t.Fatal("rowToTransaction() error = nil")
	}
	if got := err.Error(); got != `column "Valor" has cell type sheet.TextCell, want number` {
		t.Fatalf("rowToTransaction() error = %q", got)
	}
}

func TestRowToTransactionRejectsUnknownLocalizedPaymentMethod(t *testing.T) {
	row := transactionToRow(entity.Transaction{
		PaymentMethod: entity.PaymentMethodCreditCard,
	}, entity.LanguagePortugueseBrazil)
	row["Forma de pagamento"] = sheet.SelectCell("DINHEIRO")

	_, err := rowToTransaction(row, entity.LanguagePortugueseBrazil)
	if err == nil {
		t.Fatal("rowToTransaction() error = nil")
	}
	if got := err.Error(); got != `column "Forma de pagamento" has unknown payment method "DINHEIRO"` {
		t.Fatalf("rowToTransaction() error = %q", got)
	}
}

func TestLocalizedTransactionTableTitle(t *testing.T) {
	august := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	if got := localizedTransactionTableTitle(august, entity.LanguageEnglish); got != "Aug 2026" {
		t.Fatalf("English title = %q", got)
	}

	portugueseTitles := []string{
		"Jan 2026",
		"Fev 2026",
		"Mar 2026",
		"Abr 2026",
		"Mai 2026",
		"Jun 2026",
		"Jul 2026",
		"Ago 2026",
		"Set 2026",
		"Out 2026",
		"Nov 2026",
		"Dez 2026",
	}
	for index, want := range portugueseTitles {
		month := time.Date(2026, time.Month(index+1), 1, 0, 0, 0, 0, time.UTC)
		if got := localizedTransactionTableTitle(month, entity.LanguagePortugueseBrazil); got != want {
			t.Fatalf("Portuguese title for %s = %q, want %q", month.Month(), got, want)
		}
	}
}
