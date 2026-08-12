package entity

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/danielmesquitta/openfinance/internal/pkg/docutil"
)

const (
	dateTimeWithoutSeconds = "2006-01-02 15:04"
)

type Transaction struct {
	Name           string
	Category       Category
	Amount         float64
	PaymentMethod  PaymentMethod
	Date           time.Time
	Direction      TransactionDirection
	CardLastDigits *string
}

type TransactionDirection string

const (
	TransactionDirectionCredit TransactionDirection = "CREDIT"
	TransactionDirectionDebit  TransactionDirection = "DEBIT"
)

type TransactionInput struct {
	AccountType             AccountType
	Description             string
	Amount                  float64
	AmountInAccountCurrency *float64
	Date                    time.Time
	SourceCategory          string
	Direction               TransactionDirection
	PaymentMethod           *PaymentMethod
	ReceiverName            string
	ReceiverDocument        string
	CardLastDigits          *string
}

func NewTransaction(input TransactionInput) (Transaction, bool) {
	amount := input.Amount
	if input.AmountInAccountCurrency != nil && *input.AmountInAccountCurrency != 0 {
		amount = *input.AmountInAccountCurrency
	}

	transaction := Transaction{
		Amount:    math.Abs(amount),
		Date:      input.Date,
		Category:  Category(input.SourceCategory),
		Direction: input.Direction,
	}

	switch input.AccountType {
	case AccountTypeBank:
		if input.PaymentMethod == nil {
			return Transaction{}, false
		}

		transaction.PaymentMethod = *input.PaymentMethod
		transaction.Name = bankTransactionName(input)
	case AccountTypeCreditCard:
		transaction.Name = strings.TrimSpace(input.Description)
		transaction.PaymentMethod = PaymentMethodCreditCard
		transaction.CardLastDigits = input.CardLastDigits
	default:
		return Transaction{}, false
	}

	if transaction.Name == "" {
		return Transaction{}, false
	}

	return transaction, true
}

func bankTransactionName(input TransactionInput) string {
	if description := strings.TrimSpace(input.Description); description != "" {
		return description
	}

	if receiverName := strings.TrimSpace(input.ReceiverName); receiverName != "" {
		return receiverName
	}

	if input.ReceiverDocument == "" {
		return ""
	}

	document, err := docutil.MaskDocument(input.ReceiverDocument)
	if err != nil {
		return ""
	}

	return document
}

func (t Transaction) ID() string {
	return fmt.Sprintf(
		"%s:%d:%s",
		t.Name,
		int64(math.Round(t.Amount*100)), // multiply by 100 to avoid floating point precision issues
		t.Date.Format(dateTimeWithoutSeconds),
	)
}
