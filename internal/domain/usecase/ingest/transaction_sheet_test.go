package ingest

import (
	"reflect"
	"testing"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/internal/provider/sheet"
)

func TestTransactionTableOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		language entity.Language
		columns  []sheet.Column
	}{
		{
			name:     "English default",
			language: "",
			columns: []sheet.Column{
				sheet.NewTitleColumn("Name"),
				sheet.NewSelectColumn("Category", sheet.WithSelectOptions(
					sheet.NewSelectOption("Food", sheet.WithColor(entity.Red)),
					sheet.NewSelectOption("Others", sheet.WithColor(entity.Gray)),
				)),
				sheet.NewNumberColumn("Amount", sheet.WithCurrency("BRL")),
				sheet.NewSelectColumn("Payment Method", sheet.WithSelectOptions(
					sheet.NewSelectOption("BOLETO", sheet.WithColor(entity.Yellow)),
					sheet.NewSelectOption("PIX", sheet.WithColor(entity.Blue)),
					sheet.NewSelectOption("TED", sheet.WithColor(entity.Green)),
					sheet.NewSelectOption("CREDIT CARD", sheet.WithColor(entity.Purple)),
				)),
				sheet.NewTextColumn("Card Last Digits"),
				sheet.NewDateColumn("Date"),
			},
		},
		{
			name:     "Brazilian Portuguese",
			language: entity.LanguagePortugueseBrazil,
			columns: []sheet.Column{
				sheet.NewTitleColumn("Nome"),
				sheet.NewSelectColumn("Categoria", sheet.WithSelectOptions(
					sheet.NewSelectOption("Food", sheet.WithColor(entity.Red)),
					sheet.NewSelectOption("Others", sheet.WithColor(entity.Gray)),
				)),
				sheet.NewNumberColumn("Valor", sheet.WithCurrency("BRL")),
				sheet.NewSelectColumn("Forma de pagamento", sheet.WithSelectOptions(
					sheet.NewSelectOption("BOLETO", sheet.WithColor(entity.Yellow)),
					sheet.NewSelectOption("PIX", sheet.WithColor(entity.Blue)),
					sheet.NewSelectOption("TED", sheet.WithColor(entity.Green)),
					sheet.NewSelectOption("CARTÃO DE CRÉDITO", sheet.WithColor(entity.Purple)),
				)),
				sheet.NewTextColumn("Últimos dígitos do cartão"),
				sheet.NewDateColumn("Data"),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			settings := entity.IngestProfileSettings{
				Language:   test.language,
				Categories: []entity.Category{"Food", "Others"},
				ColorsByCategory: map[entity.Category]entity.Color{
					"Food": entity.Red, "Others": entity.Gray,
				},
			}

			got := resolveCreateTableOptions(transactionTableOptions(settings))
			want := sheet.CreateTableOptions{Icon: "💸", Columns: test.columns}

			if !reflect.DeepEqual(got, want) {
				t.Fatalf("definition = %#v, want %#v", got, want)
			}
		})
	}
}

func resolveCreateTableOptions(options []sheet.CreateTableOption) sheet.CreateTableOptions {
	resolved := sheet.CreateTableOptions{}
	for _, option := range options {
		if option != nil {
			option(&resolved)
		}
	}

	return resolved
}

func TestTransactionRowRoundTrip(t *testing.T) {
	t.Parallel()

	cardLastDigits := "1234"
	transaction := entity.Transaction{
		Name:           "Store",
		Category:       "Food",
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
			t.Parallel()

			row := transactionToRow(transaction, test.language)
			if row[test.columns.name] != sheet.TitleCell("Store") ||
				row[test.columns.category] != sheet.SelectCell("Food") ||
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
	t.Parallel()

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
	t.Parallel()

	columns := transactionTableLocalizationFor(entity.LanguageEnglish).columns
	row := transactionToRow(entity.Transaction{Name: "Payment made"}, entity.LanguageEnglish)

	if paymentMethod, exists := row[columns.paymentMethod]; exists {
		t.Fatalf("payment method cell = %#v, want omitted", paymentMethod)
	}

	transaction, err := rowToTransaction(row, entity.LanguageEnglish)
	if err != nil {
		t.Fatalf("rowToTransaction() error = %v", err)
	}
	if transaction.PaymentMethod != "" {
		t.Fatalf("payment method = %q, want empty", transaction.PaymentMethod)
	}
}

func TestRowToTransactionRejectsUnexpectedCellType(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	august := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	if got := localizedTransactionTableTitle(august, entity.LanguageEnglish); got != "Aug 2026" {
		t.Fatalf("English title = %q", got)
	}
	if got := localizedTransactionTableTitle(august, entity.LanguagePortugueseBrazil); got != "Agosto 2026" {
		t.Fatalf("Portuguese title = %q", got)
	}
}
