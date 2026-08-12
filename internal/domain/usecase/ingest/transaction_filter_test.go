package ingest

import (
	"testing"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

func TestShouldIgnoreTransaction(t *testing.T) {
	tests := []struct {
		name                      string
		transaction               entity.Transaction
		ignoreSamePersonTransfers bool
		want                      bool
	}{
		{
			name:        "incoming credit",
			transaction: entity.Transaction{Direction: entity.TransactionDirectionCredit},
			want:        true,
		},
		{
			name:        "investment category",
			transaction: entity.Transaction{Category: investmentCategory},
			want:        true,
		},
		{
			name:        "investment description",
			transaction: entity.Transaction{Name: "Aplicação automática"},
			want:        true,
		},
		{
			name:        "credit card bill payment",
			transaction: entity.Transaction{Name: creditCardBillPayment},
			want:        true,
		},
		{
			name:                      "configured same-person transfer",
			transaction:               entity.Transaction{Category: samePersonTransferCategory},
			ignoreSamePersonTransfers: true,
			want:                      true,
		},
		{
			name:        "allowed same-person transfer",
			transaction: entity.Transaction{Category: samePersonTransferCategory},
		},
		{
			name:        "expense",
			transaction: entity.Transaction{Name: "Market"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := shouldIgnoreTransaction(test.transaction, test.ignoreSamePersonTransfers)
			if got != test.want {
				t.Fatalf("shouldIgnoreTransaction() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestFilterTransactions(t *testing.T) {
	transactions := []entity.Transaction{
		{Name: "Market"},
		{Name: "Transfer", Category: samePersonTransferCategory},
		{Name: "Deposit", Direction: entity.TransactionDirectionCredit},
	}

	filtered := filterTransactions(transactions, false)
	if len(filtered) != 2 || filtered[0].Name != "Market" || filtered[1].Name != "Transfer" {
		t.Fatalf("filterTransactions() = %#v", filtered)
	}
}
