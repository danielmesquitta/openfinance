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
			name: "bank description and account currency amount",
			input: TransactionInput{
				AccountType:             AccountTypeBank,
				Description:             "Market",
				Amount:                  -99,
				AmountInAccountCurrency: &accountAmount,
				Date:                    date,
				PaymentMethod:           &pix,
			},
			accepted: true,
			want: Transaction{
				Name:          "Market",
				Amount:        12.34,
				Date:          date,
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
			name:     "incoming credit",
			input:    TransactionInput{Direction: TransactionDirectionCredit},
			accepted: false,
		},
		{name: "investment category", input: TransactionInput{SourceCategory: "Investments"}, accepted: false},
		{
			name:     "investment description",
			input:    TransactionInput{Description: "Aplicação automática"},
			accepted: false,
		},
		{
			name:     "card bill payment",
			input:    TransactionInput{Description: "Pagamento de fatura"},
			accepted: false,
		},
		{
			name:     "same person transfer",
			input:    TransactionInput{SourceCategory: "Same person transfer"},
			accepted: false,
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
