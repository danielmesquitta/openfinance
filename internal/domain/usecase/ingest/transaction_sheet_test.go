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

	settings := entity.IngestProfileSettings{
		Categories: []entity.Category{"Food", "Others"},
		ColorsByCategory: map[entity.Category]entity.Color{
			"Food": entity.Red, "Others": entity.Gray,
		},
	}

	got := resolveCreateTableOptions(transactionTableOptions(settings))
	want := sheet.CreateTableOptions{
		Icon: "💸",
		Columns: []sheet.Column{
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
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("definition = %#v, want %#v", got, want)
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

	row := transactionToRow(transaction)
	if row[transactionNameColumn] != sheet.TitleCell("Store") ||
		row[transactionCategoryColumn] != sheet.SelectCell("Food") ||
		row[transactionAmountColumn] != sheet.NumberCell(42.5) ||
		row[transactionPaymentMethodColumn] != sheet.SelectCell(entity.PaymentMethodCreditCard) ||
		row[transactionCardLastDigitsColumn] != sheet.TextCell("1234") ||
		row[transactionDateColumn] != sheet.DateCell(transaction.Date) {
		t.Fatalf("row = %#v", row)
	}

	got, err := rowToTransaction(row)
	if err != nil {
		t.Fatalf("rowToTransaction() error = %v", err)
	}
	if !reflect.DeepEqual(got, transaction) {
		t.Fatalf("transaction = %#v, want %#v", got, transaction)
	}
}

func TestTransactionRowEmptyCardLastDigits(t *testing.T) {
	t.Parallel()

	row := transactionToRow(entity.Transaction{})
	if row[transactionCardLastDigitsColumn] != sheet.TextCell("") {
		t.Fatalf("card cell = %#v", row[transactionCardLastDigitsColumn])
	}

	transaction, err := rowToTransaction(row)
	if err != nil {
		t.Fatalf("rowToTransaction() error = %v", err)
	}
	if transaction.CardLastDigits != nil {
		t.Fatalf("card last digits = %#v, want nil", transaction.CardLastDigits)
	}
}

func TestRowToTransactionRejectsUnexpectedCellType(t *testing.T) {
	t.Parallel()

	_, err := rowToTransaction(sheet.Row{transactionAmountColumn: sheet.TextCell("42")})
	if err == nil {
		t.Fatal("rowToTransaction() error = nil")
	}
}
