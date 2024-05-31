package meupluggyapi

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/danielmesquitta/openfinance/internal/domain/entity"
	"github.com/danielmesquitta/openfinance/pkg/formatter"
)

type listTransactionsResponse struct {
	Total      int64    `json:"total"`
	TotalPages int64    `json:"totalPages"`
	Page       int64    `json:"page"`
	Results    []result `json:"results"`
}

type result struct {
	ID                      string              `json:"id"`
	Description             string              `json:"description"`
	Amount                  float64             `json:"amount"`
	AmountInAccountCurrency *float64            `json:"amountInAccountCurrency"`
	Date                    time.Time           `json:"date"`
	Category                *string             `json:"category"`
	PaymentData             *paymentData        `json:"paymentData"`
	Type                    resultType          `json:"type"`
	CreditCardMetadata      *creditCardMetadata `json:"creditCardMetadata"`
}

type creditCardMetadata struct {
	CardNumber        *string `json:"cardNumber,omitempty"`
	TotalInstallments *int64  `json:"totalInstallments,omitempty"`
	InstallmentNumber *int64  `json:"installmentNumber,omitempty"`
}

type paymentData struct {
	Payer         *payer                `json:"payer"`
	PaymentMethod *entity.PaymentMethod `json:"paymentMethod"`
	Receiver      *payer                `json:"receiver"`
}

type payer struct {
	Name           *string         `json:"name"`
	DocumentNumber *documentNumber `json:"documentNumber"`
}

type documentNumber struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

type resultType string

const (
	Credit resultType = "CREDIT"
	Debit  resultType = "DEBIT"
)

func (c *Client) ListTransactions(
	accountID string,
	from, to time.Time,
) ([]entity.Transaction, error) {
	url := c.BaseURL

	url.Path = "/transactions"
	query := url.Query()

	query.Add("accountId", accountID)
	query.Add("pageSize", "500")

	query.Add("from", (from).Format(time.DateOnly))
	query.Add("to", (to).Format(time.DateOnly))

	req, err := http.NewRequest("GET", url.String()+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("accept", "application/json")
	req.Header.Add("X-API-KEY", c.Token)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, parseResError(res)
	}

	decoder := json.NewDecoder(res.Body)
	data := &listTransactionsResponse{}
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	transactions := []entity.Transaction{}
	for _, r := range data.Results {
		if isInvestment := (r.Category != nil && *r.Category == "Investments") ||
			(r.Description == "Aplicação RDB"); isInvestment {
			continue
		}
		if isReceivingMoney := r.Type == Credit; isReceivingMoney {
			continue
		}
		if isCreditCardBillPayment := r.
			Description == "Pagamento de fatura"; isCreditCardBillPayment {
			continue
		}

		transaction := entity.Transaction{
			Amount: math.Abs(r.Amount),
			Date:   r.Date,
		}

		if r.Category != nil {
			transaction.Category = *r.Category
		}

		accountType := entity.CreditCard
		if r.PaymentData != nil {
			accountType = entity.BankAccount
		}

		if accountType == entity.BankAccount {
			transaction.PaymentMethod = *r.PaymentData.PaymentMethod
			transaction.Description = r.Description

			if hasReceiver := (r.PaymentData.Receiver != nil); hasReceiver {
				if hasReceiverName := r.PaymentData.
					Receiver.Name != nil; hasReceiverName {
					transaction.Name = *r.PaymentData.Receiver.Name
				} else if hasReceiverDocument := r.PaymentData.
					Receiver.DocumentNumber != nil; hasReceiverDocument {
					document, _ := formatter.MaskDocument(
						r.PaymentData.Receiver.DocumentNumber.Value,
						r.PaymentData.Receiver.DocumentNumber.Type,
					)
					transaction.Name = document
				}
			}
		} else if accountType == entity.CreditCard {
			transaction.Name = r.Description
			transaction.PaymentMethod = "CREDIT CARD"
		}

		transactions = append(transactions, transaction)
	}

	return transactions, nil
}
