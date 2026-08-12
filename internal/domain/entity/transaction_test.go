package entity

import (
	"testing"
	"time"
)

func TestNewTransaction(t *testing.T) {
	t.Parallel()

	pix := PaymentMethodPix
	accountAmount := -12.34
	card := "1234"
	date := time.Date(2026, time.August, 9, 12, 30, 0, 0, time.UTC)

	tests := []struct {
		name     string
		input    TransactionInput
		accepted bool
		want     Transaction
	}{
		{
			name: "bank description, account currency amount, and source metadata",
			input: TransactionInput{
				AccountType:             AccountTypeBank,
				Description:             "Market",
				Amount:                  -99,
				AmountInAccountCurrency: &accountAmount,
				Date:                    date,
				SourceCategory:          "Groceries",
				Direction:               TransactionDirectionDebit,
				PaymentMethod:           &pix,
			},
			accepted: true,
			want: Transaction{
				Name:          "Market",
				Category:      "Groceries",
				Amount:        12.34,
				Date:          date,
				Direction:     TransactionDirectionDebit,
				PaymentMethod: PaymentMethodPix,
			},
		},
		{
			name: "bank receiver name",
			input: TransactionInput{
				AccountType:   AccountTypeBank,
				Amount:        -10,
				PaymentMethod: &pix,
				ReceiverName:  "Receiver",
			},
			accepted: true,
			want:     Transaction{Name: "Receiver", Amount: 10, PaymentMethod: PaymentMethodPix},
		},
		{
			name: "bank receiver document",
			input: TransactionInput{
				AccountType:      AccountTypeBank,
				Amount:           -10,
				PaymentMethod:    &pix,
				ReceiverDocument: "12345678000195",
			},
			accepted: true,
			want:     Transaction{Name: "12.345.678/0001-95", Amount: 10, PaymentMethod: PaymentMethodPix},
		},
		{
			name: "credit card",
			input: TransactionInput{
				AccountType:    AccountTypeCreditCard,
				Description:    "Store",
				Amount:         -20,
				CardLastDigits: &card,
			},
			accepted: true,
			want: Transaction{
				Name:           "Store",
				Amount:         20,
				PaymentMethod:  PaymentMethodCreditCard,
				CardLastDigits: &card,
			},
		},
		{
			name: "transaction filtering is deferred to the ingest use case",
			input: TransactionInput{
				AccountType:    AccountTypeBank,
				Description:    "Aplicação automática",
				SourceCategory: "Same person transfer",
				Direction:      TransactionDirectionCredit,
				PaymentMethod:  &pix,
			},
			accepted: true,
			want: Transaction{
				Name:          "Aplicação automática",
				Category:      "Same person transfer",
				Direction:     TransactionDirectionCredit,
				PaymentMethod: PaymentMethodPix,
			},
		},
		{
			name: "bank without payment method",
			input: TransactionInput{
				AccountType: AccountTypeBank,
				Description: "Store",
			},
			accepted: false,
		},
		{
			name:     "transaction without name",
			input:    TransactionInput{AccountType: AccountTypeCreditCard},
			accepted: false,
		},
		{name: "unknown account", input: TransactionInput{Description: "Store"}, accepted: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, accepted := NewTransaction(test.input)
			if accepted != test.accepted {
				t.Fatalf("accepted = %v, want %v", accepted, test.accepted)
			}
			if got != test.want {
				t.Fatalf("transaction = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestTransactionID(t *testing.T) {
	t.Parallel()

	transaction := Transaction{
		Name:   "Store",
		Amount: 10.126,
		Date:   time.Date(2026, time.August, 9, 12, 34, 56, 0, time.UTC),
	}

	if got, want := transaction.ID(), "Store:1013:2026-08-09 12:34"; got != want {
		t.Fatalf("ID() = %q, want %q", got, want)
	}
}

func TestCleanBankTransactionName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty", input: "", want: ""},
		{name: "unchanged", input: "Market", want: "Market"},
		{name: "trim spaces", input: "  Market  ", want: "Market"},
		{name: "only pipe", input: "|", want: ""},
		{name: "only spaces", input: "   ", want: ""},
		{name: "leading pipe", input: "|Market", want: "Market"},
		{name: "pipe followed by spaces", input: "| Market", want: "Market"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := cleanBankTransactionName(test.input); got != test.want {
				t.Fatalf("cleanBankTransactionName(%q) = %q, want %q", test.input, got, test.want)
			}
		})
	}
}
