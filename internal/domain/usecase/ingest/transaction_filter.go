package ingest

import (
	"slices"
	"strings"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
)

const (
	investmentCategory         = "Investments"
	samePersonTransferCategory = "Same person transfer"
	creditCardBillPayment      = "Pagamento de fatura"
)

func filterTransactions(
	transactions []entity.Transaction,
	ignoreSamePersonTransfers bool,
) []entity.Transaction {
	return slices.DeleteFunc(transactions, func(transaction entity.Transaction) bool {
		return shouldIgnoreTransaction(transaction, ignoreSamePersonTransfers)
	})
}

func shouldIgnoreTransaction(
	transaction entity.Transaction,
	ignoreSamePersonTransfers bool,
) bool {
	return transaction.Direction == entity.TransactionDirectionCredit ||
		transaction.Category == investmentCategory ||
		strings.Contains(transaction.Name, "Aplicação") ||
		transaction.Name == creditCardBillPayment ||
		ignoreSamePersonTransfers && transaction.Category == samePersonTransferCategory
}
